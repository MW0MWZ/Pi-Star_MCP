#!/bin/sh
# Build the arm64 binary, Docker image, and run the container.
# Usage: ./test/docker/run.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CONTAINER_NAME="pistar-mcp-test"

echo "==> Building arm64 binary..."
cd "$PROJECT_DIR"
make linux-arm64

echo "==> Copying binary to docker context..."
cp build/pistar-dashboard-linux-arm64 "$SCRIPT_DIR/"

echo "==> Stopping any existing container..."
docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

echo "==> Building Docker image..."
docker build -t pistar-mcp-test "$SCRIPT_DIR"

echo "==> Starting container..."
docker run -d \
    --name "$CONTAINER_NAME" \
    -p 8080:8080 \
    -p 8443:8443 \
    pistar-mcp-test

echo "==> Waiting for startup..."
sleep 2
docker logs "$CONTAINER_NAME"

echo ""
echo "============================================"
echo "  Pi-Star dashboard running:"
echo "  HTTP:  http://localhost:8080  (redirects)"
echo "  HTTPS: https://localhost:8443 (self-signed)"
echo ""
echo "  Login with any username/password."
echo ""
echo "  Stop:  docker rm -f $CONTAINER_NAME"
echo "  Logs:  docker logs -f $CONTAINER_NAME"
echo "============================================"
