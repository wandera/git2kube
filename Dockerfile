# Builder image
FROM golang:1.22.2-alpine3.19 AS builder

WORKDIR /build

ARG VERSION

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build GOMODCACHE=/go/pkg/mod GOCACHE=/root/.cache/go-build go build -v -ldflags '-w -s -X 'github.com/wandera/git2kube/cmd.Version=${VERSION}

# Runtime image
FROM alpine:3.19.1
RUN apk --no-cache add ca-certificates

RUN apk --no-cache --virtual .openssh add openssh \
    && mkdir -p /etc/ssh \
    && ssh-keyscan -t rsa github.com > /etc/ssh/ssh_known_hosts \
    && apk del .openssh

COPY --from=builder /build/git2kube /app/git2kube
WORKDIR /app

ENTRYPOINT ["./git2kube"]
