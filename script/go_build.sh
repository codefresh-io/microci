#!/bin/sh
[ -z "$DIST" ] && DIST=dist/bin

[ -z "$VERSION" ] && VERSION=$(cat VERSION)
[ -z "$BUILD_TIME" ] && BUILD_TIME=$(TZ=GMT date "+%Y-%m-%d_%H:%M_GMT")
[ -z "$VCS_COMMIT_ID" ] && VCS_COMMIT_ID=$(git rev-parse --short HEAD 2>/dev/null)
[ -z "$VCS_BRANCH_NAME" ] && VCS_BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

echo "VERSION: $VERSION"
echo "BUILD_TIME: $BUILD_TIME"
echo "VCS_COMMIT_ID: $VCS_COMMIT_ID"
echo "VCS_BRANCH_NAME: $VCS_BRANCH_NAME"

go_build() {
  [ -d "${DIST}" ] && rm -rf "${DIST:?}/*"
  [ -d "${DIST}" ] || mkdir -p "${DIST}"
  CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=${VERSION} -X main.GitCommit=${VCS_COMMIT_ID} -X main.GitBranch=${VCS_BRANCH_NAME} -X main.BuildTime=${BUILD_TIME}" \
    -v -o "${DIST}/microci" ./server/*.go
}

go_build
