# Sidecar example

- [Usage](#usage)

This example deploys a demo website, running on Python Flask, behind a caching Nginx reverse proxy. The two containers will share namespaces, so the communication happens on `localhost` between them (on loopback). Logs are streamed to make it easy to follow them.

## Usage

The easiest way to try this is:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/sidecar/install.sh | sh
```

> Pro tip: try this on the [Docker Playground](https://labs.play-with-docker.com/)

This will fetch the [install script](https://github.com/rycus86/podlike/blob/master/examples/sidecar/install.sh), and executes it. The script will clone the Git repository, and deploy the stack as `sidecar`.

You can follow the logs with:

```shell
$ docker service logs -f sidecar_pod
```

To try a few example requests, use these that return HTTP 200 OK:

```shell
$ curl -fs http://127.0.0.1:8080/ > /dev/null
$ curl -fs http://127.0.0.1:8080/page/home > /dev/null
$ curl -fs http://127.0.0.1:8080/page/specs > /dev/null
$ curl -fs http://127.0.0.1:8080/page/github > /dev/null
```

You can see the request hitting both the proxy and the app, then the proxy will cache successful responses for 5 minutes, so there shouldn't be any logs for the same URL until that expires.
