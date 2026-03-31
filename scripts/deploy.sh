#!/bin/bash
# Deploy script — pulls latest image from ECR and restarts the container
# Requires: Docker, AWS CLI, EC2 IAM role with ECR + SSM read access

set -euo pipefail

REGION="ap-southeast-1"
ECR_REPO="sherry-archive"
CONTAINER_NAME="sherry-archive"
SSM_PATH="/sherry-archive/"
ENV_FILE="/tmp/sherry-env-$(date +%s)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "==> Resolving ECR registry..."
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REGISTRY="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"

echo "==> Logging in to ECR..."
aws ecr get-login-password --region "$REGION" | \
  docker login --username AWS --password-stdin "$ECR_REGISTRY"

echo "==> Pulling latest image..."
docker pull "$ECR_REGISTRY/$ECR_REPO:latest"

echo "==> Fetching config from SSM Parameter Store..."
aws ssm get-parameters-by-path \
  --path "$SSM_PATH" \
  --with-decryption \
  --region "$REGION" \
  --query "Parameters[*].[Name,Value]" \
  --output text > /tmp/ssm_raw

while IFS=$'\t' read -r name value; do
  key="${name##*${SSM_PATH}}"
  printf '%s=%s\n' "$key" "$value"
done < /tmp/ssm_raw > "$ENV_FILE"
rm -f /tmp/ssm_raw

echo "==> Stopping old container..."
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME"   2>/dev/null || true

echo "==> Starting new container..."
docker run -d \
  --name "$CONTAINER_NAME" \
  --network host \
  --restart unless-stopped \
  --env-file "$ENV_FILE" \
  "$ECR_REGISTRY/$ECR_REPO:latest"

echo "==> Cleaning up..."
rm -f "$ENV_FILE"
docker image prune -f

echo "==> Deploy complete!"
