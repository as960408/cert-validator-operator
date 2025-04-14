# Alpine 이미지를 기반으로 시작
FROM ubuntu:latest

# 필수 패키지 설치 (curl, openssl, bash 등)
RUN apt update && apt install -y \
  bash \
  curl \
  openssl \
  && rm -rf /var/cache/apt/*

# 쉘 스크립트 복사
COPY cert_agent.sh /usr/local/bin/cert_agent.sh

# 실행 권한 부여
RUN chmod +x /usr/local/bin/cert_agent.sh

# 쉘 스크립트 실행
ENTRYPOINT ["/bin/sh", "-c", "/usr/local/bin/cert_agent.sh \"$CERT_DIR\" \"$SERVER_URL\""]
