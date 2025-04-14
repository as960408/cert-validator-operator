curl -X POST http://localhost:8080/report -vvv \
  -H "Content-Type: application/json" \
  -d '{
    "nodeName": "node1",
    "filePath": "/etc/ssl/certs/nginx.crt",
    "expiry": "2025-06-01T00:00:00Z",
    "valid": true
}'
