# Health-check example

- [Usage](#usage)

Let's assume, we have a legacy application, that we don't really want to change. It already has a *sort-of* health-check endpoint on JMX, but now we want to add this app, as the `app` component, to our monitoring, that does HTTP checks. With the [Goss](https://github.com/aelsabbahy/goss) application, running as the `goss` component, we can have a range of system level checks we can execute, to tell if the application seems to be up, and is progressing. It still cannot do JMX checks, so we also add the `exporter` component, which runs a Prometheus [JMX exporter](https://github.com/prometheus/jmx_exporter), and we'll check its output with `goss`.

The JMX and HTTP connections are possible because of the shared network namespace. The progress tests are done by checking the contents of a file on a shared volume. This example uses some mock images for demonstration, you can see what they do in their respective sub-folder, if you're interested.

## Usage

To deploy the stack straight from this repository:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/healthz/stack.yml | docker stack deploy -c - healthz
```

Then you can check the health-check output:

```shell
$ curl -s http://127.0.0.1:8080/healthz
```

Or try it on the [Docker Playground](https://labs.play-with-docker.com/?stack=https://raw.githubusercontent.com/rycus86/podlike/master/examples/healthz/stack.yml&stack_name=healthz).

