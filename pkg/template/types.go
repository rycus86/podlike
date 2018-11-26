package template

import "github.com/docker/cli/cli/compose/types"

type transformSession struct {
	WorkingDir string
	Project    *types.Config

	Args map[string]interface{}

	Configurations map[string]transformConfiguration
}

func (ts *transformSession) TODO_Set(name string, tc TC) {
	ts.registerService(name, transformConfiguration(tc))
}

type transformConfiguration struct {
	Pod         []podTemplate
	Transformer []podTemplate
	Init        []podTemplate
	Templates   []podTemplate
	Copy        []podTemplate

	Args map[string]interface{}

	Session *transformSession    `yaml:"-" mapstructure:"-"`
	Service *types.ServiceConfig `yaml:"-" mapstructure:"-"`
}

type TC transformConfiguration

type podTemplate struct {
	File   *fileTemplate
	Inline string
	Http   *httpTemplate

	IsDefault bool
}

type PT podTemplate

type fileTemplate struct {
	Path     string
	Fallback *podTemplate
}

type httpTemplate struct {
	URL      string
	Fallback *podTemplate
	Insecure bool
}

type templateVars struct {
	Service *types.ServiceConfig
	Project *types.Config

	Args map[string]interface{}
}
