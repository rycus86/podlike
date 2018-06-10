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
		Template: server.URL + "/template/testing",
		Http:     true,
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
