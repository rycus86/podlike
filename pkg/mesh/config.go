package mesh

import (
	"flag"
	"strings"
)

type Configuration struct {
	EngineConnection string
	ListenAddresses  []string
	Templates        []string
}

func parseNetworkAndAddress(a string) (string, string) {
	if strings.Contains(a, "://") {
		parts := strings.SplitN(a, "://", 2)
		return parts[0], parts[1]

	} else if len(a) > 0 && a[0] == '/' {
		return "unix", a

	}

	return "tcp", a
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func configure(args ...string) *Configuration {
	var (
		conn      string
		listen    arrayFlags
		templates arrayFlags
	)

	flags := flag.NewFlagSet("mesh", flag.PanicOnError)

	flags.StringVar(&conn, "connect", "unix:///var/run/docker.sock", "Connection to the Docker engine")
	flags.Var(&listen, "listen", "Listen address (multiple)")
	flags.Var(&templates, "template", "Template file or URL (multiple)")

	flags.Parse(args)

	if len(listen) == 0 {
		flags.Usage()
		panic("no listeners configured")
	}

	if len(templates) == 0 {
		flags.Usage()
		panic("no templates configured")
	}

	return &Configuration{
		EngineConnection: conn,
		ListenAddresses:  listen,
		Templates:        templates,
	}
}
