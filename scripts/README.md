# Podlike scripts

Command line shell wrapper scripts for common helpers of Podlike.

## podtemplate

A shell script to invoke the template generator command of the application.

```
Usage:

  podtemplate FILE [FILE...]

      Loads the given source stack yaml FILEs and
      transforms them using the templates defined in them.

  podtemplate [-h|--help]

      Prints this help string for usage.
```

By default, the `latest` version of the image is used, but can be overridden with the `PODLIKE_VERSION` environment variable.

Install it as:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/scripts/podtemplate.sh | tee /usr/local/bin/podtemplate && chmod +x /usr/local/bin/podtemplate
```

You *may* need to add `sudo` in front of `tee` and `chmod`, like:

```shell
$ curl -fsSL https://raw.githubusercontent.com/rycus86/podlike/master/scripts/podtemplate.sh | sudo tee /usr/local/bin/podtemplate && sudo chmod +x /usr/local/bin/podtemplate
```

### Sub-commands

You can also use:

```shell
$ podtemplate template ...
```

This basically ends up doing the same as above.

In the end, the generated results are used for deploying Swarm stacks, so the `podtemplate deploy` command can come in handy for automation:

```shell
$ podtemplate deploy ...  # use `docker stack deploy` args
```

This sub-command transforms the input templates, then deploys the result stack in Swarm.

```
Usage:	podtemplate deploy [OPTIONS] STACK

Loads the given templated YAML files, transforms them,
then deploys a new stack or updates an existing stack.

Options:
  -c, --compose-file strings   Path to a Compose file
      --prune                  Prune services that are no longer referenced
      --resolve-image string   Query the registry to resolve image digest and supported platforms ("always"|"changed"|"never") (default "always")
      --with-registry-auth     Send registry authentication details to Swarm agents
```

### Limitations

The template generator runs in a container, with the current working directory mounted and set as the working directory within. Because of this, paths pointing outside of this area will most probably not work, like:

```shell
$ podtemplate ../../outside/of/working/area /var/different/absolute/path
```
