# webhookd

[![Build Status](https://travis-ci.org/vision-it/webhookd.png)](https://travis-ci.org/vision-it/webhookd)


Message Broker which accepts Web Hooks from Jenkins, GitLab, GitHub, Gitea, Travis, ... and publishes them to a RabbitMQ Active Message Queue (AMQP 0-9-1).


## Building
Run `make build` or execute `go build webhookd` manually.

## Configuration
A sample configuration is provided in `webhookd.sample.json`

## License
This software is licensed under the Expat (MIT) License. Check the `LICENSE` file.
