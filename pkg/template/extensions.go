package template

import "github.com/docker/cli/cli/compose/types"

// some extensions of the internals for mesh
func (ts *transformSession) TODO_Replace(svc types.ServiceConfig, tc *TC) {
	ts.Configurations = map[string]transformConfiguration{}

	tc.Service = &svc
	ts.Project.Services = types.Services{svc} // TODO should we: append(ts.Project.Services, svc)

	ts.registerService(svc.Name, transformConfiguration(*tc))
}

type TC transformConfiguration

func NewTransformConfiguration() *TC {
	return &TC{
		Args: map[string]interface{}{},
	}
}

func (tc *TC) MergeFrom(cfg transformConfiguration) {
	for _, pod := range cfg.Pod {
		if pod.isDefault { // TODO
			continue
		}

		tc.Pod = append(tc.Pod, pod)
	}

	for _, transformer := range cfg.Transformer {
		if transformer.isDefault { // TODO
			continue
		}

		tc.Transformer = append(tc.Transformer, transformer)
	}

	tc.Init = append(tc.Init, cfg.Init...)
	tc.Templates = append(tc.Templates, cfg.Templates...)
	tc.Copy = append(tc.Copy, cfg.Copy...)

	mergeRecursively(tc.Args, cfg.Args)
}
