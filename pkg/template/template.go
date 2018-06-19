package template

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
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
		panic(fmt.Sprintf("failed to render a template : %s\n%+v", err.Error(), t))
	}

	var changed map[interface{}]interface{}
	if err := yaml.Unmarshal(buffer.Bytes(), &changed); err != nil {
		panic(fmt.Sprintf(
			"failed to decode the template result into a YAML : %s\n%s",
			err.Error(), buffer.String()))
	}

	converted, err := convertToStringKeysRecursive(changed, "")
	if err != nil {
		panic(fmt.Sprintf(
			"failed to convert the template result to map[string]interface{} recursively : %s\n%+v",
			err.Error(), changed))
	}

	return converted.(map[string]interface{})
}

func (t *podTemplate) prepareTemplate(workingDir string) *template.Template {
	name := TypeInline
	if t.File != nil {
		name = path.Base(t.File.Path)
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
			if err != nil {
				panic(fmt.Sprintf("failed to an HTTP template from %s : %s", t.Http.URL, err.Error()))
			} else {
				panic(fmt.Sprintf("failed to an HTTP template from %s : HTTP %d", t.Http.URL, resp.StatusCode))
			}

		}

	} else if t.Inline != "" {
		tmpl, err = tmpl.Parse(t.Inline)

	} else {
		tmpl, err = tmpl.ParseFiles(path.Join(workingDir, t.File.Path))
		if err != nil {
			if _, ok := err.(*os.PathError); ok && t.File.Fallback != nil {
				return t.File.Fallback.prepareTemplate(workingDir)
			} else {
				panic(fmt.Sprintf("failed to render a file template at %s : %s", t.File.Path, err.Error()))
			}
		}

	}

	if err != nil {
		panic(fmt.Sprintf("failed to render a template : %s\n%+v", err.Error(), t))
	}

	return tmpl
}
