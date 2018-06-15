package template

import (
	"bytes"
	"crypto/tls"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
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
	name := TypeInline
	if !t.Inline && t.Http == nil {
		name = path.Base(t.Template)
	}

	tmpl := template.New(name).Funcs(podTemplateFuncMap)

	var err error

	if t.Http != nil {
		var resp *http.Response

		if t.Http.Insecure {
			transport := http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			defer transport.CloseIdleConnections()

			cli := http.Client{
				Transport: &transport,
			}

			resp, err = cli.Get(t.Http.URL)

		} else {
			resp, err = http.Get(t.Http.URL)

		}

		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()

			if content, err := ioutil.ReadAll(resp.Body); err == nil {
				tmpl, err = tmpl.Parse(string(content))
			}

		} else if t.Http.Fallback != nil {
			return t.Http.Fallback.prepareTemplate(workingDir)

		} else {
			panic(err)

		}

	} else if t.Inline {
		tmpl, err = tmpl.Parse(t.Template)

	} else {
		tmpl, err = tmpl.ParseFiles(path.Join(workingDir, t.Template))

	}

	if err != nil {
		panic(err)
	}

	return tmpl
}
