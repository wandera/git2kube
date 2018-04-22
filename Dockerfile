# Builder image
FROM golang:1.10 AS builder
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep

RUN mkdir -p /go/src/github.com/WanderaOrg/git2kube/
WORKDIR /go/src/github.com/WanderaOrg/git2kube/

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY cmd/ ./cmd
COPY pkg/ ./pkg
COPY main.go .

RUN CGO_ENABLED=0 go build -o bin/git2kube


# Runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/WanderaOrg/git2kube/bin/git2kube /app/git2kube
WORKDIR /app

ENTRYPOINT ["./git2kube"]