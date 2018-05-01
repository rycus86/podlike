# Podlike

An attempt at managing co-located containers (like in a Pod in Kubernetes) mainly for services on top of Docker Swarm mode.
The general idea is the same: this container will act as a parent for the one or more children containers started as part of the *emulated* pod. Containers within this pod can use `localhost` (the loopback interface) to communicate with each other.
They can also share the same volumes, and can also see each other's PIDs, so sending UNIX signals between containers is possible.

These are always shared:

- Cgroup
- IPC namespace
- Network namespace

By default, these are also shared, but optional:

- PID namespace
- Volumes

## Use-cases

So, why would we want to do this on Docker Swarm?

1. Sidecars

You may want to always deploy an application with a supporting application, a sidecar. For example, a web application you want to be accessed only through a caching reverse proxy, or with authentication enabled, but without implementing these in the application itself.

*See also the [sidecar example](https://github.com/rycus86/podlike/tree/master/examples/sidecar)*

2. Signals

By putting containers in the same PID namespace, you send UNIX signals from one to another. Maybe an internal-only small webapp, that sends a SIGHUP to Nginx when it receives a reload request.

*See also the [signal example](https://github.com/rycus86/podlike/tree/master/examples/signal)*

3. Log collectors

With two containers sharing a local volume, you could collect and forward logs from files, that another container is writing. Maybe you have a legacy application with fixed file logging, but you'd still want to use modern log forwarders, like Fluentd.

*See also the [logging example](https://github.com/rycus86/podlike/tree/master/examples/logging)*

4. Shared volume and signals

By sharing a local volume for multiple containers, one could generate configuration for another to use, for example. Combined with singal sending, you could also ask the other app to reload it, when it is written and ready.

*See also the [volume example](https://github.com/rycus86/podlike/tree/master/examples/volume)*

## Configuration

The controller needs to run inside a Docker containers, and it needs access to the Docker engine through the API (either UNIX socket, TCP, etc.). The list of components comes from __container__ labels (not service labels). These labels need to start with `pod.container.`

For example:

```yaml
version: '3.5'
services:

  pod:
    image: rycus86/podlike
    command: -logs
    labels:
      # sample app with HTML responses
      pod.container.app: |
        image: rycus86/demo-site
        environment:
          - HTTP_HOST=127.0.0.1
          - HTTP_PORT=12000
      # caching reverse proxy
      pod.container.proxy: |
        image: nginx:1.13.10
      # copy the config file for the proxy
      pod.copy.proxy: /var/conf/nginx.conf:/etc/nginx/conf.d/default.conf
    configs:
      - source: nginx-conf
        target: /var/conf/nginx.conf
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    ports:
      - 8080:80

configs:
  nginx-conf:
    file: ./nginx.conf
    # the actual configuration proxies requests from port 80 to 12000 on localhost
```

Or as a simple container for testing:

```shell
$ docker run --rm -it --name podtest                      \
    -v /var/run/docker.sock:/var/run/docker.sock:ro       \
    -v $PWD/nginx.conf:/etc/nginx/conf.d/default.conf:ro  \
    --label pod.container.app='
image: rycus86/demo-site
environment:
  - HTTP_HOST=127.0.0.1
  - HTTP_PORT=12000'                                      \
    --label pod.container.proxy='
image: nginx:1.13.10'                                     \
    -p 8080:80                                            \
    rycus86/podlike -logs
```

See the [examples folder](https://github.com/rycus86/podlike/tree/master/examples) with more, small example stacks.

## Dragons!

This project is very much work in progress (see below). Even with all the tasks done, this will never enable full first-class support for pods on Docker Swarm the way Kubernetes does. Still, it might be useful for small projects or specific deployments.

I'm not yet sure how the components' containers will interfere with Swarm scheduling, resource allocation, etc. Memory limits are honored, but the components are limited to the controller's limits at most. Memory reservation is allowed on the components if you really want to, but comes with a warning. If you set the reservation on the controller, the cgroup should take note of this for you for all the containers.

The current implementation also needs the Docker API connection, usually the engine's UNIX socket as a volume, which will be available to each of the components as well (unless volume sharing is disabled with `-volumes=false`.

Some Swarm features are also *hacked around*, for example configs and secrets can be available to the controller container, but I haven't found easy way to share those with the component containers. These configuration can be copied at component startup, by adding a `pod.copy.<name>=/source/file/in/controller:/dest/file/in/component` label on the controller. It does mean, that on every startup or restart, these will be copied again, just be aware. Swarm service labels are also not available on container, and the controller doesn't assume it's running on a Swarm manager node, so we need to use container labels here, which is a bit of a shame.

Component reaping is done on a best-effort basis, killing the controller could leave you with zombie containers. With the components placed within the controller's cgroup, plus with PID sharing enabled, this is probably somewhat mitigated, but you could still potentialy end up having containers using memory and CPU after the controller dies. The components are also started with auto-remove, so getting information about them post-mortem might prove difficult.

## Work in progress

Some of the open tasks are:

- [ ] Support for most settings for the components *based on Composefile keys*
- [ ] List the supported keys, and gate on having this list up to date in the README
- [ ] CPU limits and reservation
- [ ] The stop grace period of the components should be smaller than the controller's
- [ ] The stop grace period is not visible on containers, only on services
- [ ] Swarm service labels are not visible on containers, only on services
- [ ] With volume sharing enabled, the Docker socket will be visible to all components, when visible to the controller
- [ ] Consider adding a `pause` container
- [ ] Example implementations for [composite containers patterns](https://kubernetes.io/blog/2015/06/the-distributed-system-toolkit-patterns)
- [x] Support for memory limits
- [x] Note on how memory reservation *may* affect Swarm scheduling
- [x] Support for healthchecks
- [x] Small usage examples
- [x] Sharing Swarm secrets and configs with the components - copy on start
- [x] Do we want logs collected from the components - now optional

## License

MIT
