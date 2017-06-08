# Stage 1
FROM debian:stretch as builder

ENV GOPATH=/go

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    golang-go

COPY src/webhookd/ /go/src/webhookd/

WORKDIR /go/src/webhookd/

RUN go get -t -d ./...
RUN CGO_ENABLED=0 GOOS=linux go build webhookd


# Stage 2
FROM scratch

EXPOSE 8080

COPY --from=builder /go/src/webhookd/webhookd /webhookd
CMD ["/webhookd"]
