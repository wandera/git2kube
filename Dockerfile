# Builder image
FROM golang:1.11 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v

# Runtime image
FROM alpine:3.8
RUN apk --no-cache add ca-certificates

RUN apk --no-cache --virtual .openssh add openssh \
    && mkdir -p /etc/ssh \
    && ssh-keyscan -t rsa github.com > /etc/ssh/ssh_known_hosts \
    && apk del .openssh

COPY --from=builder /build/git2kube /app/git2kube
WORKDIR /app

ENTRYPOINT ["./git2kube"]