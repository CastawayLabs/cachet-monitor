# build stage
FROM golang:alpine AS build-env
ENV WORKDIR /go/src/github.com/castawaylabs/cachet-monitor
RUN mkdir -p ${WORKDIR}
ADD . ${WORKDIR}
WORKDIR ${WORKDIR}
RUN apk add --no-cache ca-certificates git
RUN go get -v github.com/Sirupsen/logrus \
      github.com/docopt/docopt-go \
      github.com/mitchellh/mapstructure \
      gopkg.in/yaml.v2 \
      github.com/miekg/dns
RUN cd cli && go build -o /app/cachet-monitor

# final stage
FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build-env /app/cachet-monitor /app/
WORKDIR /app
RUN chmod a+x cachet-monitor
ENTRYPOINT ["/app/cachet-monitor"]
