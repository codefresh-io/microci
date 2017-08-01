#!/bin/sh
# Generate test coverage statistics for Go packages.
#
# Works around the fact that `go test -coverprofile` currently does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#
# Usage: script/coverage [--html]
#
#     --html  Additionally create HTML report and open it in browser
#
[ -z "$COVER" ] && COVER=.cover
profile="$COVER/cover.out"
mode=atomic

OS=$(uname)
race_flag="-race"
if [ "$OS" = "Linux" ]; then
  # check Alpine - alpine does not support race test
  if [ -f "/etc/alpine-release" ]; then 
    race_flag=""
  fi
fi

generate_cover_data() {
  [ -d "${COVER}" ] && rm -rf "${COVER:?}/*"
  [ -d "${COVER}" ] || mkdir -p "${COVER}"

  for pkg in "$@"; do
    f="${COVER}/$(echo $pkg | tr / -).cover"
    tout="${COVER}/$(echo $pkg | tr / -)_tests.out"
    go test $race_flag -covermode="$mode" -coverprofile="$f" "$pkg" | tee "$tout"
  done

  echo "mode: $mode" >"$profile"
  grep -h -v "^mode:" "${COVER}"/*.cover >>"$profile"
}

show_cover_report() {
  go tool cover -${1}="$profile" -o "${COVER}/coverage.html"
}

generate_cover_data "$(go list ./... | grep -v vendor)"
show_cover_report html
