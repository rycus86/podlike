# Templates

Podlike comes with a built-in template processor, to help transforming services and their tasks in Docker Swarm stacks into a *"pod"*, a set of co-located and tightly coupled containers. Similar stacks, or similar types of applications in a stack, could often benefit from decorating the tasks with the same components, with only slightly different configuration. For example, [sidecars](https://github.com/rycus86/podlike/tree/master/examples/sidecar) or [service meshes](https://github.com/rycus86/podlike/tree/master/examples/service-mesh) usually need the same component consistently deployed with the applications themselves, and we'd probably want changes to these components done in a single place, and applied to all services (or a set of them) at the same time.

Using templates gives you a flexible way to define these components, and allows you to reuse the components accross stacks and services. Templates generate parts of a YAML [Compose file](https://docs.docker.com/compose/compose-file/) using Go's [text/template package](https://golang.org/pkg/text/template/). The templates to use are defined directly in the stack YAML files with extension fields, so the configuration lives within them, and can be versioned/changed/rolled out with the same workflows as the original stack.

## Configuration

The example YAML snippet below demonstrates the use of the Podlike-related [extension fields](https://docs.docker.com/compose/compose-file/#extension-fields) the template generator will look for.

```yaml
version: '3.5'

x-anchors:

  - &tracing-component
    templates/components/tracing.yml

  - &proxy-component
    templates/components/proxy.yml

x-podlike:
  example:
    pod:
      - templates/pod.yml
      - http: https://templates.store.local/pods/example.yml
      - inline:
          pod:
            labels:
              swarm.service.label: templated-for-{{ .Service.Name }}
    transformer:
      - templates/transformer.yml
      # or http or inline
    templates:
      - templates/components/sidecar.yml
      - templates/components/logger.yml
      - *tracing-component
      - *proxy-component
      # or http or inline
    copy:
      - inline:
          sidecar: /var/conf/sidecar.conf:/etc/sidecar/conf.d/default.conf
      # or http or from file
    args:
      ExposedHttpPort: 8080
      ExtraEnvVars:
        - DEBUG=false
      # any other arguments you'd need available for the templates
  args:
    GlobalVariable: example
    # any other arguments you'd need available for the templates

services:

  example:
    image: example/app:0.1.2
    volumes:
      - log-data:/var/logs/example
    x-podlike:
      pod:
        # same as above
      transformer:
        # same as above
      templates:
        # same as above
      copy:
        # same as above
      args:
        # same as above

volumes:
  log-data:
    name: log-data-for-{{.Task.ID}}
    labels:
      com.github.rycus86.podlike.volume-ref: shared-log-folder
```

Let's unpack the example above, and look at the different extension places.

## Top-level extension

Extension fields at the root level of a stack YAML are supported since Compose schema version `3.4`, and are simply ignored by a `docker stack deploy`. Podlike can use the `x-podlike` top-level extension field to define templates *per service*, matching the service name, plus any additional `args` to make available globally to the templates used within the stack.

For each service, we can define `pod` templates, `transformer` templates, `templates` for the additional components, `copy` configurations, and additional `args` available for templates used with this service. The additional arguments are merged with any global `args`.

```yaml
# top-level extension
x-podlike:
  svc1:
    pod:
    transformer:
    templates:
    copy:
    args:
  svc2:
    templates:
    args:
  args:
```

Every field is optional to use, and you can use a single template or a list of them for `pod`, `transformer`, `templates` or `copy`. If multiple templates are given for a single type, they will be merged together, in order - see more details below at [Template merging](#template-merging).

The example above would define templates to use on the `svc1` and `svc2` services, plus specific arguments for each service, as well as additional global arguments. See which template is used for what below, but first, let's have a quick overview of what types of parameters they accept.

### Template definition types

All 4 types of definitions accept either a single item, or a list of items. An item can be:

1. A simple string

This points to a template file or an HTTP(S) address.

```yaml
x-podlike:
  example:
    pod:
      - templates/pod.yml
    transformer:
      - file:
          path: local.template.yml
          fallback:
            inline: InlineFallbackIfFileNotAvailable
    templates:
      - https://template.srv.local/addon.yml
```

File templates can be given with a simple *string* pointing to a file, or as a mapping with a mandatory `path` field, and an optional `fallback` property if loading the file fails.

2. An inline template mapping

This uses the given string as the template text, or the YAML string marshalled from the given mapping.

```yaml
x-podlike:
  example:
    pod:
      - inline: |
          image: sample/{{ .Service.Name }}:{{ .Args.ImageTag }}
          labels:
            format: string
      - inline:
          labels:
            given: as.mapping
```

3. An HTTP(S) URL to the template

This fetches the template from the given URL, and uses the response content as the template text.

```yaml
x-podlike:
  example:
    pod:
      - http: https://my.templates.local/pods/sample.yml
    transformer:
      - http:
          url: https://maybe.insecure.local/transformer/sample.yml
          insecure: true
      - http:
          url: http://template.cache.local/addons/cache.yml
          fallback:
            file:
              path: cached/local.copy.yml
              fallback:
                inline: |
                  main:
                    image: sensible/defaults
```

As shown above, the value of the `http` property can be a simple URL *string*, or a mapping with a mandatory `url` field, plus an optional `insecure` property to disable SSL certificate validation, and a `fallback` field to specify the template to try if loading from HTTP fails.

### Controller templates

The templates listed under the `pod` key are used to construct the new Swarm service definition for the *controller*. This is allowed to produce a Swarm compatible service mapping, e.g. `deploy`, `configs`, `secrets`, etc. are OK.

If omitted, a default template is used to generate the `image` property pointing to `rycus86/podlike` with the same version as the template generator. The default also adds a volume mapping for the Docker engine socket at `/var/run/docker.sock` for convenience, and enables streaming logs from the components using the `-logs` Podlike flag.

If there is at least one template given, the template engine only makes sure there is an `image` defined, with the same rules as above, plus it adds volume for the Docker engine socket.

```yaml
x-podlike:
  example:
    pod:
      inline:
        pod:
          image: forked/podlike:{{ .Args.Version }}
          deploy:
            replicas: 3
    args:
      Version: 0.1.2
```

The name of the root property in the generated string doesn't matter, it will be replaced by the actual name of the service as given in the stack YAML. The template engine also copies over most of the properties from the original service definition, unless they are added by the templates, see these in the `mergedPodKeys` in the [merge.go](https://github.com/rycus86/podlike/blob/master/pkg/template/merge.go) source file.

Each of the templates given here must output a YAML compatible string with a single root property.

### Main component templates

The `transformer` templates generate the Compose-compatible *component* definition for the main component, that is the original image defined in the stack YAML in most cases, with its selected properties.

```yaml
x-podlike:
  example:
    transformer:
      inline: |
        main:
          environment:
            - EXTRA_VARS={{ .Args.ExtraEnv }}
          {{ if .Service.ReadOnly }}
          read_only: true
          {{ end }}
    args:
      ExtraEnv: some-env-var
```

It no templates given, a default one will copy over the `image` property from the original service definition, plus a fair bit of other properties are added automatically, defined by `mergedTransformerKeys` in the [merge.go](https://github.com/rycus86/podlike/blob/master/pkg/template/merge.go) source file.

Each of the templates given here must output a YAML compatible string with a single root property. Most of the [v2 Compose file](https://docs.docker.com/compose/compose-file/compose-file-v2/) properties are allowed, with the exceptions listed on the main project [README](https://github.com/rycus86/podlike/blob/master/README.md#unsupported-properties). The name of the result *component* will be the root key of the first template. Root keys defined by any other templates will be ignored, and converted automatically to the one defined by the first. The example above would use `main`, the default is `app`.

### Additional component templates

The templates listed under `templates` can define any number of *components*. These are meant to generate the Compose-compatible definitions of the containers to couple the main component with.

```yaml
x-podlike:
  example:
    templates:
      - templates/sidecar.yml
      - templates/service-discovery.yml
      - templates/tracing.yml
      - inline:
          tracing:
            mem_limit: 64m
      - inline:
          tracing:
            environment:
              HTTP_PORT: '{{ .Args.Tracing.Http.Port }}'
    args:
      Tracing:
        Http:
          Port: 12345
```

As with the other types, the templates are processed in the same order as they are defined in the YAML, and any common properties are merged in together. In the example above, the `templates/tracing.yml` template could define a component with the `tracing` name, then the last two templates would add in the `mem_limit` property, if not defined by the previous template already, plus the `environment` variables would also contain `HTTP_PORT`.

The names of the components come from root properties of the result YAML, after merging all the template outputs together.

### Copy templates

Podlike allows copying files from the *controller* container into the *component* containers before they start, and the `copy` templates can define the mappings for these.

```yaml
x-podlike:
  example:
    copy:
      - inline:
          proxy: '/shared/proxy.conf:/var/conf/proxy/default.conf'
      - inline:
          logging:
            - /shared/logging.conf:/var/conf/logger/settings.properties
            - /shared/proxy.logging:/var/conf/logger/conf.d/proxy.conf
```

Each template needs to output a mapping of service name to copy configurations. The copy items will be converted into a `string` slice of `<source>:<target>` paths, but accepts a `<source>: <target>` mapping, or a single string as well. The lists generated by all the `copy` templates will be then merged into a single list, and put on the *controller* definition.

## Template merging

As mentioned above, each type of templates can use multiple source to generate the final markup, and they can output the same properties for the same component with different settings. Single-valued properties are going to be ignored if redefined, but *slices* and *maps* are merged together. A prime example of these would be `environment` variables or `labels`.

```yaml
x-podlike:
  example:
    transformer:
      - inline:
          environment:
            - HTTP_PROXY=my.local.proxy:8091
          labels:
            inline.label: sample
      - inline:
          environment:
            ADDED: 'new key, and is added'
            HTTP_PROXY: 'ignored as already defined'
            # note that `- HTTP_PROXY=override` would have been added
            # because at this point the template engine wouldn't assume it's
            # a key-value pair as a string, only when it sees that it can be a mapping
          labels:
            inline.label:      ignored
            additional.label:  added
```

The merging logic works on a best-effort basis to merge items of the same property together, even if they are of different types. It can:

- Merge items of a *map* into another *map*
- Merge items of a *slice* of `key=value` pairs into a *map*
- Merge items of a *slice* into another *slice*
- Merge items of a *map* into a *slice* after converting it to a *map* as `key=value` pairs
- Add a *string* into a *slice*

See the implementation in the [merge.go](https://github.com/rycus86/podlike/blob/master/pkg/template/merge.go) file, and also the tests for these cases in the [merge_test.go](https://github.com/rycus86/podlike/blob/master/pkg/template/merge_test.go) file.

## Service-level extension

Besides top-level extension fields, the template engine also supports per-service extensions with the same `x-podlike` name. This currently works by removing the property and its children from the YAML after reading the configuration from them.

> With Compose schema version `3.7`, the service-level extension fields are going to be [supported](https://github.com/docker/cli/pull/1097) as well, but until then having these makes the YAML invalid for a plain `docker stack deploy` command.

The configuration is the same as it is for the top-level field, with the exception that the service name does not have to be defined as it is inferred from the service name.

```yaml
version: '3.5'
services:
  
  example:
    image: sample/application
    x-podlike:
      pod:
      transformer:
      templates:
      copy:
      args:
```

If the same service has any configuration in the top-level `x-podlike` field as well, then those are merged into the service-level configuration following the rules above. For example:

```yaml
version: '3.5'
services:
  
  example:
    image: sample/application
    x-podlike:
      templates:
        - templates/first.yml
        - templates/second.yml

x-podlike:
  example:
    templates:
      - templates/third.yml
```

The `args` are also merged the same way, in the order of:

1. Service-level arguments
2. Per-service arguments from the top-level extension
3. Global arguments from the top-level extension

This allows you to define default values for arguments globally, then override then per service.

## Using YAML anchors

If multiple services in a single stack share similar templating configuration, YAML anchors could help reduce some of the duplication. For example:

```yaml
version: '3.5'

x-podlike-templates:  # the name of this does not matter
  - &default-pod
    pod:
      inline:
        image: forked/podlike
        command: -logs -pids=false

  - &sidecar-template
    inline:
      sidecar:
        image: sample/sidecar

  - &logging-template
    inline:
      logging:
        image: sample/logger
        command: -input {{ .Args.LogFile }}

services:
 
  service-one:
    image: sample/svc1
    x-podlike:
      <<: *default-pod
      templates:
        - <<: *sidecar-template
        - <<: *logging-template
      args:
        LogFile: /var/logs/service.one.log

  service-two:
    image: sample/svc2
    x-podlike:
      <<: *default-pod
      templates:
        - <<: *logging-template
      args:
        LogFile: /var/logs/service.two.log
```

This is particularly useful when using inline templates. A better approach would be sharing the templates through files, or serving them up on HTTP.

## Template variables and functions

When rendering the templates, the following variables are available to them:

- `Service`: the Swarm service definition as the [Docker cli package](https://github.com/docker/cli/blob/master/cli/compose/types/types.go) has it
- `Args`: the merged *map* of the service arguments, with the global `args` added in as described above

There are also additional template functions available, on top of the [built-in ones](https://golang.org/pkg/text/template/#hdr-Functions):

- `yaml <obj>`: returns the YAML string representation of an object
- `indent <num> <str>`: indents every line of `<str>` by `<num>` spaces
- `empty <obj>`: returns *true* if the array/slice/map is empty
- `notEmpty <obj>`: returns *true* if the array/slice/map is not empty
- `contains <s> <t>`: returns *true* if `<t>` contains `<s>`
- `startsWith <s> <t>`: return *true* if `<t>` starts with `<s>`
- `replace <old> <new> <n> <s>`: replaces `<old>` with `<new>` `<n>` times in `<s>`

An example template using them could look like this:

```
sidecar:
  image: sidecars/{{ .Args.Sidecar.Current.Image }}:{{ .Args.Sidecar.Current.Version }}
{{ if notEmpty .Service.Ports }}
  {{ with $port := index .Service.Ports 0 }}
  command: --listen {{ $port.Target }}
  {{ end }}
{{ else }}
  command: --listen 8080
{{ end }}
  labels:
{{ range $key, $value := .Service.Labels }}
  {{ if $key | startsWith "sidecar." }}
    {{ with $label := $key | replace "sidecar." "" -1 }}
{{ printf "%s: %s" $label $value | indent 4 }}
    {{ end }}
  {{ end }}
{{ end }}
```

If given an input like this:

```yaml
version: '3.5'
services:

  app:
    image: sample/app:1.1.0
    ports:
      - 9090:3000
      - 15000:8080
    labels:
      com.xyz.label: value
      sidecar.metrics.port: 15000
      sidecar.ping.endpoint: /v2/ping
    x-podlike:
      pod:
        inline:
          pod:
            labels:  # avoid copying labels from the service
      transformer:
        inline:
          app:
            labels:
              com.xyz.system: sample-app
      templates:
        - templates/sidecar/v1.yml
      args:
        Sidecar:
          Current:
            Image: scapp
            Version: 3.4.2
```

The generated result would look like this:

```yaml
version: "3.5"
services:
  app:
    image: rycus86/podlike:latest
    labels:
      pod.component.app: |
        image: sample/app:1.1.0
        labels:
          com.xyz.system: sample-app
      pod.component.sidecar: |
        command:
        - --listen
        - "3000"
        image: sidecars/scapp:3.4.2
        labels:
          metrics.port: "15000"
          ping.endpoint: /v2/ping
    ports:
    - mode: ingress
      target: 3000
      published: 9090
      protocol: tcp
    - mode: ingress
      target: 8080
      published: 15000
      protocol: tcp
    volumes:
    - type: bind
      source: /var/run/docker.sock
      target: /var/run/docker.sock
      read_only: true
networks: {}
volumes: {}
secrets: {}
configs: {}
```

## Usage

The template engine is part of the main application, so it can easily be invoked through Docker:

```shell
$ docker run --rm -it                \
        -v $PWD:/workspace:ro        \
        -w /workspace                \
        rycus86/podlike              \
        template <file> [<file>...]
```

This shares the current directory (and its sub-directories) and generates the final YAML using the given input stack YAML files. Alternatively, you can pipe the source stack to the standard input of the container, if you only use inline templates:

```shell
$ cat source.yml | docker run --rm -i rycus86/podlike template -
# notice the missing --tty option
```

To make these easier, there is a script to do this for you, called `podtemplate`. Have a look into its [documentation](https://github.com/rycus86/podlike/tree/master/scripts) for installation and usage information.
