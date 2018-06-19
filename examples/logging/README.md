# Log collection example

- [Usage](#usage)

This stack shows an example of how one container could collect the logs of another container, if the application inside it logs to files. This could be useful if there is an existing application, that *needs* to log to files, but you'd still want to collect and forward them with something like Fluentd or Logstash.

In this small example, the `logger` container writes the logs to a shared volume, and the `tail` container is reading them, dumping the lines to the output of the "pod".

## Usage

To deploy the stack straight from this repository:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/logging/stack.yml | docker stack deploy -c - logging
```

You can follow the logs with this:

```shell
$ docker service logs -f logging_pod
```

Or try it on the [Docker Playground](https://labs.play-with-docker.com/?stack=https://raw.githubusercontent.com/rycus86/podlike/master/examples/logging/stack.yml&stack_name=logging).

