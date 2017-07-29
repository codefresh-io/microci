#
# ----- Go Builder Image ------
#
FROM golang:1.8-alpine3.6 AS builder

# gox - Go cross compile tool
# github-release - Github Release and upload artifacts
# go-junit-report - convert Go test into junit.xml format
RUN apk add --no-cache git bash curl && \
    go get -v github.com/aktau/github-release && \
    go get -v github.com/jstemmer/go-junit-report

# set working directory
RUN mkdir -p /go/src/github.com/codefresh-io/microci
WORKDIR /go/src/github.com/codefresh-io/microci

# copy sources
COPY . .

# set entrypoint to bash
ENTRYPOINT ["/bin/bash"]

# run test and calculate coverage
RUN VERSION=$(cat VERSION) script/coverage.sh
# upload coverage reports to Codecov.io: pass CODECOV_TOKEN as build-arg
ARG CODECOV_TOKEN
RUN bash -c "bash <(curl -s https://codecov.io/bash) -t ${CODECOV_TOKEN}"

# build microci binary
RUN VERSION=$(cat VERSION) script/go_build.sh


#
# ------ MicroCI runtime image ------
#
FROM alpine:3.6

# add root certificates
RUN apk add --no-cache ca-certificates
# add user:group
RUN addgroup microci && adduser -s /bin/bash -D -G microci microci

ARG GOSU_VERSION=1.10
ARG GOSU_SHA_256=5b3b03713a888cee84ecbf4582b21ac9fd46c3d935ff2d7ea25dd5055d302d3c

RUN apk add --no-cache --virtual .gosu-deps curl && \
    curl -o /tmp/gosu-amd64 -LS  "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-amd64" && \
    echo "${GOSU_SHA_256}  gosu-amd64" > /tmp/gosu-amd64.sha256 && \
    cd /tmp && sha256sum -c gosu-amd64.sha256 && \
    mv /tmp/gosu-amd64 /usr/local/bin/gosu && \
    chmod +x /usr/local/bin/gosu && \
    gosu nobody true && \
    rm -rf /tmp/* && \
    apk del .gosu-deps

COPY --from=builder /go/src/github.com/codefresh-io/microci/dist/bin/microci /usr/bin/microci
COPY docker_entrypoint.sh /
RUN chmod +x /docker_entrypoint.sh

ENTRYPOINT ["/docker_entrypoint.sh"]
CMD ["microci", "--help"]

ARG GH_SHA=dev
LABEL org.label-schema.vcs-ref=$GH_SHA \
      org.label-schema.vcs-url="https://github.com/codefresh-io/microci"
