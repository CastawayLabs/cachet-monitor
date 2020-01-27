# build stage
FROM golang:alpine AS build-env
RUN set -eux; \
    apk add --no-cache --virtual .build-deps \
    git gcc libc-dev;
ENV GO111MODULE on
WORKDIR /go/main
ADD go.mod go.sum ./
RUN go mod download
# WORKDIR /go/src/status-monitor
ADD main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o main main.go
