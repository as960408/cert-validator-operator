#!/bin/sh

# 환경 변수에서 디렉터리 경로와 서버 URL을 받음
CERT_DIR="$1"
SERVER_URL="$2"

# 인증서 유효성 검사 함수
check_cert_expiry() {
  CERT_FILE=$1
  if [ ! -f "$CERT_FILE" ]; then
    echo "파일을 찾을 수 없습니다: $CERT_FILE"
    exit 1
  fi

  # 인증서 만료일자 추출 (예: 2025-06-01T00:00:00Z)
  EXPIRY_DATE=$(openssl x509 -enddate -noout -in "$CERT_FILE" | cut -d= -f2)

  # 만약 EXPIRY_DATE가 비어 있거나 잘못된 값이라면 에러 메시지 출력
  if [ -z "$EXPIRY_DATE" ]; then
    echo "인증서의 만료일을 추출할 수 없습니다: $CERT_FILE"
    exit 1
  fi

  # 날짜 포맷을 변환 (예: 2025-06-01T00:00:00Z 형식으로 변환)
  EXPIRY_DATE_FORMATTED=$(date -d "$EXPIRY_DATE" --utc +%Y-%m-%dT%H:%M:%SZ)

  # 유효성 체크 (만료일이 현재 날짜보다 이후인지 확인)
  EXPIRY_TIMESTAMP=$(date --date="$EXPIRY_DATE_FORMATTED" +%s)
  CURRENT_TIMESTAMP=$(date +%s)

  if [ "$EXPIRY_TIMESTAMP" -gt "$CURRENT_TIMESTAMP" ]; then
    VALID=true
  else
    VALID=false
  fi

  # 결과 리턴 (유효성, 만료일, 파일경로)
  echo "$EXPIRY_DATE_FORMATTED $VALID $CERT_FILE"
}

# 1시간마다 반복 실행
while true; do
  # 주어진 경로 내의 모든 .crt 파일에 대해 처리
  for CERT_FILE in "$CERT_DIR"/*.crt; do
    if [ -f "$CERT_FILE" ]; then
      # 인증서 유효성 체크
      RESULT=$(check_cert_expiry "$CERT_FILE")
      EXPIRY_DATE=$(echo "$RESULT" | awk '{print $1}')
      VALID=$(echo "$RESULT" | awk '{print $2}')
      CERT_FILE_PATH=$(echo "$RESULT" | awk '{print $3}')

      # JSON 데이터 생성
      JSON_PAYLOAD=$(cat <<EOF
{
  "nodeName": "$NODE_NAME",
  "filePath": "$CERT_FILE_PATH",
  "expiry": "$EXPIRY_DATE",
  "valid": $VALID
}
EOF
      )

      # JSON_PAYLOAD 출력 (디버깅용)

      # 서버로 데이터 전송
      curl -X POST "$SERVER_URL/report" \
        -H "Content-Type: application/json" \
        -d "$JSON_PAYLOAD"
    fi
  done
      echo "Send the following JSON payload to $SERVER_URL/report"


  # 1시간 대기
  sleep 3600
done

