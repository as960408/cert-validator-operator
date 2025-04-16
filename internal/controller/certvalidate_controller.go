package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/as960408/cert-validator-operator/api/v1alpha1"
)

type CertValidateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CertValidateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var certValidate corev1alpha1.CertValidate
	if err := r.Get(ctx, req.NamespacedName, &certValidate); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled CertValidate", "name", req.Name)
	return ctrl.Result{}, nil
}

func (r *CertValidateReconciler) createOrUpdateDaemonSet(ctx context.Context, namespace, nodeSelectorStr, certDir string) error {
	dsName := "cert-agent"
	labels := map[string]string{"app": "cert-agent"}

	nodeSelector := map[string]string{}
	if nodeSelectorStr != "" {
		parts := strings.Split(nodeSelectorStr, "=")
		if len(parts) == 2 {
			nodeSelector[parts[0]] = parts[1]
		}
	}

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080/report"
	}

	tolerations := []corev1.Toleration{
		{Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
		{Operator: corev1.TolerationOpExists, Key: "CriticalAddonsOnly"},
		{Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute},
	}

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dsName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Tolerations:  tolerations,
					NodeSelector: nodeSelector,
					Containers: []corev1.Container{
						{
							Name:  "agent",
							Image: "as960408/cert-agent:0.4",
							Env: []corev1.EnvVar{
								{Name: "CERT_DIR", Value: certDir},
								{Name: "SERVER_URL", Value: serverURL},
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "pki",
									MountPath: certDir,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "pki",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: certDir,
								},
							},
						},
					},
				},
			},
		},
	}

	found := &appsv1.DaemonSet{}
	err := r.Get(ctx, client.ObjectKey{Name: dsName, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return r.Create(ctx, daemonSet)
	} else if err != nil {
		return err
	}

	daemonSet.ResourceVersion = found.ResourceVersion
	return r.Update(ctx, daemonSet)
}

// HTTP 서버에서 CertAgent로부터 데이터를 수신
func (r *CertValidateReconciler) startServer() {
	http.HandleFunc("/report", r.handleReport)
	fmt.Println("[HTTP] Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}

func (r *CertValidateReconciler) handleReport(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var report CertReport
	if err := json.Unmarshal(body, &report); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx)

	name := generateName(report.NodeName, report.FilePath)

	var existing corev1alpha1.CertValidate
	err = r.Get(ctx, client.ObjectKey{Name: name, Namespace: "default"}, &existing)

	if errors.IsNotFound(err) {
		cr := &corev1alpha1.CertValidate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Spec: corev1alpha1.CertValidateSpec{
				NodeName: report.NodeName,
				FilePath: report.FilePath,
				Expiry:   report.Expiry,
				Valid:    report.Valid,
			},
		}
		err = r.Create(ctx, cr)
		if err != nil {
			logger.Error(err, "Failed to create CertValidate")
			http.Error(w, "Create failed", http.StatusInternalServerError)
			return
		}
		logger.Info("Created CertValidate", "name", name)
		w.WriteHeader(http.StatusCreated)
		return
	} else if err != nil {
		http.Error(w, "Failed to check CR", http.StatusInternalServerError)
		return
	}

	existing.Spec = corev1alpha1.CertValidateSpec{
		NodeName: report.NodeName,
		FilePath: report.FilePath,
		Expiry:   report.Expiry,
		Valid:    report.Valid,
	}
	err = r.Update(ctx, &existing)
	if err != nil {
		logger.Error(err, "Failed to update CertValidate")
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	logger.Info("Updated CertValidate", "name", name)
	w.WriteHeader(http.StatusOK)
}

func generateName(nodeName, filePath string) string {
	cleanPath := strings.ReplaceAll(filePath, "/", "-")
	cleanPath = strings.Trim(cleanPath, "-")
	return fmt.Sprintf("%s-%s", nodeName, cleanPath)
}

// 오퍼레이터 시작 시 DaemonSet 배포
func (r *CertValidateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	go r.startServer()

	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		nodeSelector := os.Getenv("NODE_SELECTOR")
		certDir := os.Getenv("CERT_DIR")
		if certDir == "" {
			certDir = "/etc/kubernetes/pki"
		}

		if err := r.createOrUpdateDaemonSet(ctx, "default", nodeSelector, certDir); err != nil {
			fmt.Printf("❌ Failed to create cert-agent DaemonSet: %v\n", err)
		} else {
			fmt.Println("✅ cert-agent DaemonSet created or updated successfully.")
		}
		return nil
	}))

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CertValidate{}).
		Complete(r)

}

// +kubebuilder:rbac:groups=core.certwatcher.io,resources=certvalidates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.certwatcher.io,resources=certvalidates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.certwatcher.io,resources=certvalidates/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=pods,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=apps,resources=services,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete

type CertReport struct {
	NodeName string `json:"nodeName"`
	FilePath string `json:"filePath"`
	Expiry   string `json:"expiry"`
	Valid    bool   `json:"valid"`
}
