#!/usr/bin/env sh
set -eu

: "${AWS_ACCOUNT_ID:?AWS_ACCOUNT_ID を設定してください}"
: "${AWS_REGION:?AWS_REGION を設定してください}"
: "${IMAGE_TAG:?IMAGE_TAG を設定してください}"
: "${BACKEND_PUBLIC_URL:?BACKEND_PUBLIC_URL を設定してください}"
: "${AUTH_ISSUER:?AUTH_ISSUER を設定してください}"

AUTH_CLIENT_ID="${AUTH_CLIENT_ID:-commerce-frontend}"
REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
BACKEND_IMAGE="${REGISTRY}/commerce-lab-backend:${IMAGE_TAG}"
FRONTEND_IMAGE="${REGISTRY}/commerce-lab-frontend:${IMAGE_TAG}"

docker build -t "${BACKEND_IMAGE}" ./backend
docker build \
  --build-arg NEXT_PUBLIC_API_BASE_URL="${BACKEND_PUBLIC_URL}" \
  --build-arg NEXT_PUBLIC_AUTH_ISSUER="${AUTH_ISSUER}" \
  --build-arg NEXT_PUBLIC_AUTH_CLIENT_ID="${AUTH_CLIENT_ID}" \
  -t "${FRONTEND_IMAGE}" ./frontend

docker push "${BACKEND_IMAGE}"
docker push "${FRONTEND_IMAGE}"

echo "BACKEND_IMAGE=${BACKEND_IMAGE}"
echo "FRONTEND_IMAGE=${FRONTEND_IMAGE}"
