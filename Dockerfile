FROM golang:1.12-alpine as builder

RUN apk add --no-cache make curl git gcc musl-dev linux-headers

ADD . /go/src/github.com/CastawayLabs/cachet-monitor
RUN cd /go/src/github.com/CastawayLabs/cachet-monitor/cli && go get && go build -o cachet-monitor

# Copy into a second stage container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/CastawayLabs/cachet-monitor/cli/cachet-monitor /usr/local/bin/

ENTRYPOINT ["cachet-monitor"]