FROM golang

ADD . /go/src/github.com/castawaylabs/cachet-monitor
RUN go install github.com/castawaylabs/cachet-monitor

ENTRYPOINT /go/bin/cachet-monitor