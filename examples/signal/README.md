# Signal example

This simple example starts two Python processes. One of them has a signal handler for `SIGHUP`, and prints a hello when signalled. It also writes its own PID to a shared volume, for the second container to read. This one will then send the signal to this PID every second.

## Usage

To deploy the stack straight from this repository:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/signal/stack.yml | docker stack deploy -c - signal
```

You can follow the logs with this:

```shell
$ docker service logs -f signal_pod
```

Or try it on the [Docker Playground](https://play-with-docker.com/?stack=https://raw.githubusercontent.com/rycus86/podlike/master/examples/signal/stack.yml&stack_name=signal).

