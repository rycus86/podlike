# Podlike

An attempt at managing co-located containers (like in a Pod in Kubernetes) mainly for services on top of Docker Swarm mode.
The general idea is the same: this container will act as a parent for the one or more children containers started as part of the "emulated" pod.

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

2. Signals

By putting containers in the same PID namespace, you send UNIX signals from one to another. Maybe an internal-only small webapp, that sends a SIGHUP to Nginx when it receives a reload request.

3. Shared temporary storage

By sharing a local volume for multiple containers, one could generate configuration for another to use, for example. Combined with singal sending, you could also ask the other app to reload it, when it is written and ready.

## Configuration

The controller needs to run inside a Docker containers, and it needs access to the Docker engine through the API (either UNIX socket, TCP, etc.). The list of components comes from __container__ labels (not service labels). These labels need to start with `pod.container.`

For example:

```yaml
version: '3.5'
services:

  example:
    image: rycus86/podlike
    labels:
      - pod.container.srv-1: |
          image: alpine
          command: nc -p 99 -v -le sleep 3
      - pod.container.cli-2: |
          image: alpine
          command: sh -c "sleep 1 && nc -z localhost 99 && echo OK"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

Or as a simple container for testing:

```shell
docker run --name pltest --rm -it \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    --label abc=12 \
    --label pod.container.srv-1='image: alpine
command: nc -p 99 -v -le sleep 3' \
    --label pod.container.cli-2='image: alpine
command: sh -c "sleep 1 && nc -z localhost 99 && echo OK"' \
    podlike
```

*Note:* the `rycus86/podlike` image is not published to Docker Hub yet.

## Work in progress

Some of the open tasks are:

- [ ] Small usage examples
- [ ] Support for many-many more settings you can configure for the components' containers
- [ ] CPU and Memory limit and reservation distribution within the pod
- [ ] How does memory limit and reservation on the components affect Swarm scheduling
- [ ] Do we want logs collected from the components
- [ ] The stop grace period of the components should be smaller than the controller's
- [ ] The stop grace period is not visible on containers, only on services
- [ ] Swarm service labels are not visible on containers, only on services
- [ ] With volume sharing enabled, the Docker socket will be visible to all components, when visible to the controller
- [ ] Sharing Swarm secrets and configs with the components

## License

MIT