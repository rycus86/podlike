package template

import (
	"github.com/docker/cli/cli/compose/types"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRender_File(t *testing.T) {
	tmplFile, err := ioutil.TempFile("", "template")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmplFile.Name())

	if _, err := tmplFile.WriteString(`
component:
  image: sample/{{ .Service.Labels.CompName }}:{{ .Args.TargetTag }}
`); err != nil {
		t.Fatal(err)
	}
	if err := tmplFile.Close(); err != nil {
		t.Fatal(err)
	}

	tmpl := podTemplate{
		Template: tmplFile.Name(),
	}

	rendered := tmpl.render(&transformConfiguration{
		Service: &types.ServiceConfig{
			Labels: types.Labels{
				"CompName": "testing",
			},
		},
		Args: map[string]interface{}{
			"TargetTag": "0.1.2",
		},
		Session: &transformSession{},
	})

	if comp, ok := rendered["component"]; !ok {
		t.Error("Root key not found")
	} else if mComp, ok := comp.(map[string]interface{}); !ok {
		t.Error("Invalid root key")
	} else if image, ok := mComp["image"]; !ok {
		t.Error("Image key not found")
	} else if image != "sample/testing:0.1.2" {
		t.Error("Invalid image value found")
	}
}

func TestRender_Inline(t *testing.T) {
	tmpl := podTemplate{
		Template: `
component:
  image: sample/{{ .Service.Labels.CompName }}:{{ .Args.TargetTag }}
`,
		Inline: true,
	}

	rendered := tmpl.render(&transformConfiguration{
		Service: &types.ServiceConfig{
			Labels: types.Labels{
				"CompName": "testing",
			},
		},
		Args: map[string]interface{}{
			"TargetTag": "0.1.2",
		},
		Session: &transformSession{},
	})

	if comp, ok := rendered["component"]; !ok {
		t.Error("Root key not found")
	} else if mComp, ok := comp.(map[string]interface{}); !ok {
		t.Error("Invalid root key")
	} else if image, ok := mComp["image"]; !ok {
		t.Error("Image key not found")
	} else if image != "sample/testing:0.1.2" {
		t.Error("Invalid image value found")
	}
}

func TestRender_Http(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/template/testing" {
			t.Fatal("Invalid template request")
		}

		w.WriteHeader(200)
		w.Write([]byte(`
component:
  image: sample/{{ .Service.Labels.CompName }}:{{ .Args.TargetTag }}
`))
	}))
	defer server.Close()

	tmpl := podTemplate{
		Http: &httpTemplate{
			URL: server.URL + "/template/testing",
		},
	}

	rendered := tmpl.render(&transformConfiguration{
		Service: &types.ServiceConfig{
			Labels: types.Labels{
				"CompName": "testing",
			},
		},
		Args: map[string]interface{}{
			"TargetTag": "0.1.2",
		},
		Session: &transformSession{},
	})

	if comp, ok := rendered["component"]; !ok {
		t.Error("Root key not found")
	} else if mComp, ok := comp.(map[string]interface{}); !ok {
		t.Error("Invalid root key")
	} else if image, ok := mComp["image"]; !ok {
		t.Error("Image key not found")
	} else if image != "sample/testing:0.1.2" {
		t.Error("Invalid image value found")
	}
}

func TestRender_HttpWithSelfSignedCert(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/template/testing/tls" {
			t.Fatal("Invalid template request")
		}

		w.WriteHeader(200)
		w.Write([]byte(`
proxy:
  image: sample/proxy
  command: serve --8080
`))
	}))
	defer server.Close()

	tmpl := podTemplate{
		Http: &httpTemplate{
			URL:      server.URL + "/template/testing/tls",
			Insecure: true,
		},
	}

	rendered := tmpl.render(&transformConfiguration{
		Service: &types.ServiceConfig{},
		Args:    map[string]interface{}{},
		Session: &transformSession{},
	})

	if comp, ok := rendered["proxy"]; !ok {
		t.Error("Root key not found")
	} else if mComp, ok := comp.(map[string]interface{}); !ok {
		t.Error("Invalid root key")
	} else if image, ok := mComp["image"]; !ok {
		t.Error("Image key not found")
	} else if image != "sample/proxy" {
		t.Error("Invalid image value found")
	}
}

func TestRender_HttpWithFallback(t *testing.T) {
	tmpl := podTemplate{
		Http: &httpTemplate{
			URL: "http://127.0.0.1:65001/not/found",
			Fallback: &podTemplate{
				Template: `
component:
  image: sample/http:fallback
`,
				Inline: true,
			},
		},
	}

	rendered := tmpl.render(&transformConfiguration{
		Service: &types.ServiceConfig{},
		Args:    map[string]interface{}{},
		Session: &transformSession{},
	})

	if comp, ok := rendered["component"]; !ok {
		t.Error("Root key not found")
	} else if mComp, ok := comp.(map[string]interface{}); !ok {
		t.Error("Invalid root key")
	} else if image, ok := mComp["image"]; !ok {
		t.Error("Image key not found")
	} else if image != "sample/http:fallback" {
		t.Error("Invalid image value found")
	}
}

func TestRender_WithFuncs(t *testing.T) {
	tmpl := podTemplate{
		Template: `
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
`,
		Inline: true,
	}

	rendered := tmpl.render(&transformConfiguration{
		Service: &types.ServiceConfig{
			Labels: types.Labels{
				"sidecar.label.a": "val-a",
				"sidecar.label.b": "val-b",
			},
			Ports: []types.ServicePortConfig{
				{Target: 9090, Published: 80},
				{Target: 9091, Published: 15000},
			},
		},
		Args: map[string]interface{}{
			"Sidecar": map[string]interface{}{
				"Current": map[string]interface{}{
					"Image":   "example",
					"Version": "0.2.3",
				},
			},
		},
		Session: &transformSession{},
	})

	if comp, ok := rendered["sidecar"]; !ok {
		t.Error("Root key not found")
	} else if mComp, ok := comp.(map[string]interface{}); !ok {
		t.Error("Invalid root key")
	} else if image, ok := mComp["image"]; !ok {
		t.Error("Image key not found")
	} else if image != "sidecars/example:0.2.3" {
		t.Error("Invalid image value found")
	} else if command, ok := mComp["command"]; !ok {
		t.Error("Command key not found")
	} else if command != "--listen 9090" {
		t.Error("Invalid command found")
	} else if labels, ok := mComp["labels"]; !ok {
		t.Error("Labels not found")
	} else if mLabels, ok := labels.(map[string]interface{}); !ok {
		t.Error("Invalid labels found")
	} else if v, ok := mLabels["label.a"]; !ok || v != "val-a" {
		t.Error("Invalid label value found")
	} else if v, ok := mLabels["label.b"]; !ok || v != "val-b" {
		t.Error("Invalid label value found")
	}
}
