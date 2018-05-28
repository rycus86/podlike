# Examples

You can find simple examples in the subfolders. Most of them can be run with a simple `docker stack deploy -c stack.yml example`, then follow the logs with `docker service logs -f example_pod`, to see what is happening. Each subfolder has its own *README* though for more detailed instructions.

> Pro tip: try the examples on the [Docker Playground](https://labs.play-with-docker.com/)

If you want to try it on your machine, make sure that Swarm is initialized, and run `docker swarm init` if it isn't.

> See a more detailed explanation of the examples at [https://blog.viktoradam.net/2018/05/24/podlike-example-use-cases/](https://blog.viktoradam.net/2018/05/24/podlike-example-use-cases/)
