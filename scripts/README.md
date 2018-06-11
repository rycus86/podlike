# Podlike scripts

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

*You may need to add `sudo` in front of `tee` and `chmod`.*
