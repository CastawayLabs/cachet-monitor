#!/bin/bash

set -e

if [ "$1" == "test" ]; then
  reflex -r '\.go$' -s -d none -- sh -c 'go test ./...'
fi

reflex -r '\.go$' -s -d none -- sh -c 'go build -o ./cachet-monitor ./cli/ && ./cachet-monitor -c config.yml'
exit 0
