#!/bin/sh
[ -z "$DIST" ] && DIST=dist/bin

[ -z "$VERSION" ] && VERSION=$(cat VERSION)
[ -z "$BUILD_TIME" ] && BUILD_TIME=$(TZ=GMT date "+%Y-%m-%d_%H:%M_GMT")
[ -z "$GIT_COMMIT" ] && GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null)
[ -z "$GIT_BRANCH" ] && GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

echo "VERSION: $VERSION"
echo "BUILD_TIME: $BUILD_TIME"
echo "GIT_COMMIT: $GIT_COMMIT"
echo "GIT_BRANCH: $GIT_BRANCH"

go_build() {
  [ -d "${DIST}" ] && rm -rf "${DIST:?}/*"
  [ -d "${DIST}" ] || mkdir -p "${DIST}"
  CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.GitBranch=${GIT_BRANCH} -X main.BuildTime=${BUILD_TIME}" \
    -v -o "${DIST}/microci" ./server/*.go
}

go_build
