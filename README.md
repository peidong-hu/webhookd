# webhookd

[![Build Status](https://travis-ci.org/vision-it/webhookd.png)](https://travis-ci.org/vision-it/webhookd)


Message Broker which accepts Web Hooks from Jenkins, GitLab, GitHub, Gitea, Travis, ... and publishes them to a RabbitMQ Active Message Queue (AMQP 0-9-1).


## Building
Run `make build` or execute `go build webhookd` manually.

## Configuration
A sample configuration is provided in `webhookd.sample.json`

## Debugging
This repo contains a program called 'listen' which will read the same configuration file as 'webhookd' (because it uses the same credentials and options for the message queue) and act as a consumer on the other side of the message queue. You can build it with `make listen` or `go build listen`.
It may also serve as an example on how to implement a consumer for the message queue in Go.

## License
This software is licensed under the Expat (MIT) License. Check the `LICENSE` file.
