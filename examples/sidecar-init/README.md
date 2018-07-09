# Sidecar with init component example

- [Usage](#usage)

This example deploys a demo website, running on Python Flask, behind a caching Nginx reverse proxy. The two containers will share namespaces, so the communication happens on `localhost` between them (on loopback). Logs are streamed to make it easy to follow them.

The Nginx configuration is set up by an [init component](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/). You can either use `depends_on` for this if the container supports health checks, or the `pod.init.components` definition. In the latter case, the init containers will run to completion one-by-one before any other components are started.

## Usage

To deploy the stack straight from this repository:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/sidecar-init/stack-depends-on.yml | docker stack deploy -c - mesh
```

> Pro tip: try this on the [Docker Playground](https://labs.play-with-docker.com/?stack=https://raw.githubusercontent.com/rycus86/podlike/master/examples/sidecar-init/stack-depends-on.yml&stack_name=mesh)

You can also try the templated version, once you have the `podtemplate` script [installed](https://github.com/rycus86/podlike/tree/master/scripts):

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/sidecar-init/stack-init.yml | podtemplate deploy -c - mesh
```

You can follow the logs with:

```shell
$ docker service logs -f mesh_site
```

To try a few example requests, use these that return HTTP 200 OK:

```shell
$ curl -fs http://127.0.0.1:8080/ > /dev/null
$ curl -fs http://127.0.0.1:8080/page/home > /dev/null
$ curl -fs http://127.0.0.1:8080/page/specs > /dev/null
$ curl -fs http://127.0.0.1:8080/page/github > /dev/null
```

You can see the request hitting both the proxy and the app, then the proxy will cache successful responses for 5 minutes, so there shouldn't be any logs for the same URL until that expires.
