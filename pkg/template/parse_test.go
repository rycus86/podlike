package template

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestParse_Simple(t *testing.T) {
	verifyParseResults(t, `
version: '3.5'
services:
  test:
    image: test/validate

x-podlike:
  test:
    pod: simple.yml
`, verifyParsedService("test", func(c transformConfiguration) bool {
		return len(c.Pod) == 1 && c.Pod[0].File != nil && c.Pod[0].File.Path == "simple.yml"
	}))
}

func TestParse_FileOrHttp(t *testing.T) {
	verifyParseResults(t, `
version: '3.5'
services:
  test:
    image: test/validate

x-podlike:
  test:
    templates: 
      - from-file.yml
      - http://remote.srv/tmpl.yml
      - https://secure.srv/addon.yml
`, verifyParsedService("test", func(c transformConfiguration) bool {
		return len(c.Templates) == 3 &&
			c.Templates[0].File != nil && c.Templates[0].File.Path == "from-file.yml" &&
			c.Templates[1].Http != nil && c.Templates[1].Http.URL == "http://remote.srv/tmpl.yml" &&
			c.Templates[2].Http != nil && c.Templates[2].Http.URL == "https://secure.srv/addon.yml"
	}))
}

func TestParse_Http(t *testing.T) {
	verifyParseResults(t, `
version: '3.5'
services:
  test:
    image: test/validate

x-podlike:
  test:
    pod:
      http: https://direct.value
    transformer:
      http:
        url: http://insecure.target
        insecure: true
    templates:
      - http:
          url: http://with.fallback
          fallback:
            inline: 'InlineTemplate'
      - http:
          url: https://insecure.fallback
          insecure: true
          fallback: from/file.yml
`,
		verifyParsedService("test", func(c transformConfiguration) bool {
			return len(c.Pod) == 1 && c.Pod[0].Http != nil &&
				c.Pod[0].Http.URL == "https://direct.value" &&
				c.Pod[0].Http.Insecure == false &&
				c.Pod[0].Http.Fallback == nil
		}),
		verifyParsedService("test", func(c transformConfiguration) bool {
			return len(c.Transformer) == 1 && c.Transformer[0].Http != nil &&
				c.Transformer[0].Http.URL == "http://insecure.target" &&
				c.Transformer[0].Http.Insecure == true &&
				c.Transformer[0].Http.Fallback == nil
		}),
		verifyParsedService("test", func(c transformConfiguration) bool {
			return len(c.Templates) == 2 &&
				c.Templates[0].Http != nil && c.Templates[1].Http != nil &&

				c.Templates[0].Http.URL == "http://with.fallback" &&
				c.Templates[0].Http.Insecure == false &&
				c.Templates[0].Http.Fallback != nil &&
				c.Templates[0].Http.Fallback.Inline == "InlineTemplate" &&

				c.Templates[1].Http.URL == "https://insecure.fallback" &&
				c.Templates[1].Http.Insecure == true &&
				c.Templates[1].Http.Fallback != nil &&
				c.Templates[1].Http.Fallback.Inline == "" &&
				c.Templates[1].Http.Fallback.File != nil &&
				c.Templates[1].Http.Fallback.File.Path == "from/file.yml"
		}))
}

func TestParse_Inline(t *testing.T) {
	verifyParseResults(t, `
version: '3.5'
services:
  test:
    image: test/validate

x-podlike:
  test:
    transformer:
      inline:
        main:
          image: sample/transformer
    templates:
      - inline: |
          extra:
            image: sample/templated
`,
		verifyParsedService("test", func(c transformConfiguration) bool {
			return len(c.Transformer) == 1 && strings.Contains(c.Transformer[0].Inline, "image: sample/transformer")
		}),
		verifyParsedService("test", func(c transformConfiguration) bool {
			return len(c.Templates) == 1 && strings.Contains(c.Templates[0].Inline, "image: sample/templated")
		}))
}

type parseAssert func(c map[string]transformConfiguration) bool

func verifyParseResults(t *testing.T, input string, asserts ...parseAssert) {
	f, err := ioutil.TempFile("", "podlike-parse-test")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f.WriteString(input)

	session := newSession(f.Name())

	for idx, assert := range asserts {
		if !assert(session.Configurations) {
			t.Errorf("Parse result assertion #%d failed for input:\n%s", idx+1, input)
		}
	}
}

func verifyParsedService(service string, assert func(transformConfiguration) bool) parseAssert {
	return func(configs map[string]transformConfiguration) bool {
		if config, ok := configs[service]; !ok {
			return false
		} else {
			return assert(config)
		}
	}
}
