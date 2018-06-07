package template

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"path"
	"text/template"
)

func (t *podTemplate) render(tc *transformConfiguration) map[string]interface{} {
	var (
		buffer  = new(bytes.Buffer)
		allArgs = map[string]interface{}{}
	)

	mergeRecursively(allArgs, tc.Args)
	mergeRecursively(allArgs, tc.Session.Args)

	if err := t.prepareTemplate(tc.Session.WorkingDir).Execute(buffer, templateVars{
		Service: tc.Service,
		Project: tc.Session.Project,
		Args:    allArgs,
	}); err != nil {
		panic(err)
	}

	var changed map[interface{}]interface{}
	if err := yaml.Unmarshal(buffer.Bytes(), &changed); err != nil {
		panic(err)
	}

	converted, err := convertToStringKeysRecursive(changed, "")
	if err != nil {
		panic(err)
	}

	return converted.(map[string]interface{})
}

func (t *podTemplate) prepareTemplate(workingDir string) *template.Template {
	name := "inline"
	if !t.Inline {
		name = path.Base(t.Template)
	}

	tmpl := template.New(name).Funcs(template.FuncMap{
		"yaml":     yamlFunc,
		"indent":   indentFunc,
		"empty":    emptyFunc,
		"notEmpty": notEmptyFunc,
	})

	var err error

	if t.Inline {
		tmpl, err = tmpl.Parse(t.Template)
	} else {
		tmpl, err = tmpl.ParseFiles(path.Join(workingDir, t.Template))
	}
	if err != nil {
		panic(err)
	}

	return tmpl
}
