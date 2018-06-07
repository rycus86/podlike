package template

import (
	"strings"
)

func PrintTemplatedStack(parameters ...string) {
	if len(parameters) == 0 || parameters[0] == "-h" || parameters[0] == "--help" {
		println(help())

	} else {
		println(Transform(parameters...))

	}
}

func help() string {
	return strings.TrimSpace(`
Podlike template
----------------

This command preprocesses a set of Compose files for Swarm stacks,
turns the services marked up in them into "pods", and prints the
transformed YAML in a 'docker stack deploy -c - <stack-name>' compatible format.
Services can be transformed if they have an 'x-podlike' definition, and
can define templates for generating the controller, the main component,
any additional components and the copy configuration sections.

Find more detailed information at https://github.com/rycus86/podlike

Usage:

  podlike template FILE [FILE...]

      Loads the given source stack yaml FILEs and
      transforms them using the templates defined in them.

  podlike template [-h|--help]

      Prints this help string for usage.
`)
}
