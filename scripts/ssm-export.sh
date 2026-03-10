#!/bin/bash
# Fetch SSM parameters and export as environment variables
# Usage: source scripts/ssm-export.sh

REGION="ap-southeast-1"
SSM_PATH="/sherry-archive/"

aws ssm get-parameters-by-path \
  --path "$SSM_PATH" \
  --with-decryption \
  --region "$REGION" \
  --query "Parameters[*].[Name,Value]" \
  --output text > /tmp/ssm_raw

while IFS=$'\t' read -r name value; do
  key="${name##*${SSM_PATH}}"
  export "${key}=${value}"
done < /tmp/ssm_raw

COUNT=$(wc -l < /tmp/ssm_raw)
rm -f /tmp/ssm_raw

echo "==> Exported $COUNT parameters from SSM"
