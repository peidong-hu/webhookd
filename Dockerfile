# Stage 1
FROM debian:stretch as builder

ENV GOPATH=/go

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    golang-go \
    make

COPY . /go/src/webhookd/

WORKDIR /go/src/webhookd/

RUN CGO_ENABLED=0 GOOS=linux \
    make build-dep webhookd listener


# Stage 2
FROM scratch

EXPOSE 8080

WORKDIR /

COPY --from=builder /go/src/webhookd/webhookd /webhookd
COPY --from=builder /go/src/webhookd/listener /listener
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["/webhookd"]
