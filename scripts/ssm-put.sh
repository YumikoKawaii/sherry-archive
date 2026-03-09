#!/bin/bash
# Populate SSM Parameter Store for sherry-archive
# Fill in the values below, then run: bash scripts/ssm-put.sh

set -euo pipefail

REGION="ap-southeast-1"

put() {
  aws ssm put-parameter \
    --name "/sherry-archive/$1" \
    --value "$2" \
    --type "$3" \
    --region "$REGION" \
    --overwrite
  echo "  set $1"
}

echo "==> Writing parameters to SSM..."

# Database
put DB__HOST              ""        String
put DB__PORT              "5432"    String
put DB__USER              ""        String
put DB__PASSWORD          ""        SecureString
put DB__NAME              ""        String
put DB__SSL_MODE          "require" String
put DB__MIGRATIONS_SOURCE "/migrations" String

# JWT
put JWT__ACCESS_SECRET        "" SecureString
put JWT__REFRESH_SECRET       "" SecureString
put JWT__ACCESS_TOKEN_EXPIRY  "15m"  String
put JWT__REFRESH_TOKEN_EXPIRY "168h" String

# S3
put S3__REGION         "ap-southeast-1" String
put S3__BUCKET         "sherry-archive" String
put S3__PRESIGN_EXPIRY "1h"             String
put AWS_ACCESS_KEY_ID     "" SecureString
put AWS_SECRET_ACCESS_KEY "" SecureString

# Redis (container name on Docker network)
put REDIS__ADDR "redis:6379" String
put REDIS__DB   "0"          String

echo "==> Done."
