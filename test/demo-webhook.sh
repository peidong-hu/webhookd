#!/bin/sh

HOST=${HOST:-localhost}
PORT=${PORT:-8080}
PAYLOAD=${PAYLOAD:-"{\"repository\":\"my-repo\", \"branch\":\"my-branch\", \"author\":\"me\", \"message\":\"Hello World!\"}"}

curl --data "payload=$PAYLOAD" "http://${HOST}:${PORT}/webhooks/test"
