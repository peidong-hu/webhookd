# webhookd

[![Build Status](https://travis-ci.org/vision-it/webhookd.png)](https://travis-ci.org/vision-it/webhookd)


Message Broker which accepts Web Hooks from various services and publishes them to a RabbitMQ Active Message Queue (AMQP 0-9-1).

## Webhooks
Implemented webhooks:

- [X] GitHub
- [X] Travis
- [ ] GitLab
- [ ] Gitea
- [ ] Jenkins


## Building
Run `make build` or execute `go build webhookd` manually.

Alternatively, you can build your own Docker image with the supplied `Dockerfile`. Please note this Dockerfile uses the new [multi-stage builds feature](https://docs.docker.com/engine/userguide/eng-image/multistage-build/) and therefore requires at least version 17.05 of Docker.

## Configuration
A sample configuration is provided in `webhookd.sample.json`

## Debugging
This repo contains a program called 'listen' which will read the same configuration file as 'webhookd' (because it uses the same credentials and options for the message queue) and act as a consumer on the other side of the message queue. You can build it with `make listen` or `go build listen`.
It may also serve as an example on how to implement a consumer for the message queue in Go.

Additionally, 'webhookd' has an integrated basic web hook for testing. It can be enabled on the command-line and accepts web hooks of the `x/www-url-form-encoded` type with a JSON payload. An example for this can be found in the `test/test-webhook.sh` script.

## License
This software is licensed under the Expat (MIT) License. Check the `LICENSE` file.
