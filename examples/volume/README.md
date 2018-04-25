# Volume example

This example starts an Nginx container, and a very basic Python webserver. This accepts any `HTTP GET` requests on port `5000`, and when it does, it regenerates the Nginx configuration on the shared volume, and sends a reload signal to it. The generated configuration will get Nginx to respond to any `HTTP GET` requests on port `8000` (on the host), with the path part of the request and the date of the last configuration write.

## Usage

To deploy the stack straight from this repository:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/examples/volume/stack.yml | docker stack deploy -c - volume
```

You can follow the logs with this:

```shell
$ docker service logs -f volume_pod
```

To trigger the configuration updates, you can try:

```shell
$ curl -fs http://127.0.0.1:5000/reload
OK, Nginx reloaded with pid: 24
My pid is: 15
$ curl -fs http://127.0.0.1:8000/
Reloaded from /reload at Wed Apr 25 17:25:46 2018
$ curl -fs http://127.0.0.1:5000/update
OK, Nginx reloaded with pid: 24
My pid is: 15
$ curl -fs http://127.0.0.1:8000/
Reloaded from /update at Wed Apr 25 17:26:15 2018
```

Or try it on the [Docker Playground](https://play-with-docker.com/?stack=https://raw.githubusercontent.com/rycus86/podlike/master/examples/volume/stack.yml&stack_name=volume).

