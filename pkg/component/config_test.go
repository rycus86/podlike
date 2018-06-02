package component

import (
	"bytes"
	"github.com/rycus86/podlike/pkg/convert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestDeserialization(t *testing.T) {
	item, err := deserialize(`
image: alpine
command: echo "hello"
environment:
  - ENV=environment
working_dir: /var/app
labels:
  label.one: first label
  label.two: second label
tty: true
stop_signal: SIGINT
stop_grace_period: 1m30s
`)

	if err != nil {
		t.Fatal("Failed to deserialize config:", err)
	}

	if item.Image != "alpine" {
		t.Error("Wrong image name:", item.Image)
	}

	if item.Command != "echo \"hello\"" {
		t.Error("Wrong command:", item.Command)
	}

	if item.WorkingDir != "/var/app" {
		t.Error("Wrong working directory:", item.WorkingDir)
	}

	environment, err := convert.ToStringSlice(item.Environment)
	if err != nil {
		t.Error("Wrong environemnt variables:", err)
	}

	if len(environment) != 1 || environment[0] != "ENV=environment" {
		t.Error("Wrong environment variables:", item.Environment)
	}

	labels, err := convert.ToStringToStringMap(item.Labels)
	if err != nil {
		t.Error("Wrong labels:", err)
	}

	if len(labels) != 2 {
		t.Error("Wrong number of labels:", item.Labels)
	}

	if value, ok := labels["label.one"]; !ok || value != "first label" {
		t.Error("Wrong labels:", item.Labels)
	}

	if value, ok := labels["label.two"]; !ok || value != "second label" {
		t.Error("Wrong labels:", item.Labels)
	}

	if !item.Tty {
		t.Error("Wrong tty value:", item.Tty)
	}

	if item.StopSignal != "SIGINT" {
		t.Error("Wrong stop signal:", item.StopSignal)
	}

	if item.StopGracePeriod.Seconds() != 90 {
		t.Error("Wrong stop grace period:", item.StopGracePeriod)
	}
}

func TestSlices(t *testing.T) {
	item, err := deserialize(`
image: sample
entrypoint:
  - setup
  - script
command: ["sh", "-c", "ping && pong"]
`)

	if err != nil {
		t.Fatal("Failed to deserialize config:", err)
	}

	if slice, ok := item.Entrypoint.([]interface{}); !ok {
		t.Error("Wrong entrypoint:", item.Entrypoint)
	} else if slice[0] != "setup" || slice[1] != "script" {
		t.Error("Wrong entrypoint:", item.Entrypoint)
	}

	if slice, ok := item.Command.([]interface{}); !ok {
		t.Error("Wrong command:", item.Command)
	} else if slice[0] != "sh" || slice[1] != "-c" || slice[2] != "ping && pong" {
		t.Error("Wrong command:", item.Command)
	}
}

func TestHealthCheck(t *testing.T) {
	item, err := deserialize(`
image: sample
healthcheck:
  test: curl -fs localhost:8000/
  interval: 1m30s
  timeout: 10s
  retries: 3
  start_period: 40s
`)

	if err != nil {
		t.Fatal("Failed to deserialize config:", err)
	}

	if item.Healthcheck == nil {
		t.Fatal("Failed to deserialize healthcheck:")
	}

	if item.Healthcheck.Test != "curl -fs localhost:8000/" {
		t.Error("Wrong test:", item.Healthcheck.Test)
	}

	if item.Healthcheck.Interval.Seconds() != 90 {
		t.Error("Wrong interval:", item.Healthcheck.Interval)
	}

	if item.Healthcheck.Timeout.Seconds() != 10 {
		t.Error("Wrong timeout:", item.Healthcheck.Timeout)
	}

	if item.Healthcheck.Retries != 3 {
		t.Error("Wrong retries:", item.Healthcheck.Retries)
	}

	if item.Healthcheck.StartPeriod.Seconds() != 40 {
		t.Error("Wrong start period:", item.Healthcheck.StartPeriod)
	}
}

func contains(item string, list []string) bool {
	for _, test := range list {
		if item == test {
			return true
		}
	}

	return false
}

func mapContains(item string, m map[string]string) bool {
	parts := strings.SplitN(item, "=", 2)
	value, ok := m[parts[0]]
	return ok && value == parts[1]
}

func TestEnvAndLabelsAsSliceOrMap(t *testing.T) {
	var (
		item *Component
		err  error

		environment []string
		labels      map[string]string
	)

	item, err = deserialize(`
image: as_slice
environment:
  - ONE=1
  - TWO=2
labels:
  - label.one=l-one
  - label.two=l-two
`)
	if err != nil {
		t.Error("Failed to deserialize:", err)
	}

	environment, err = convert.ToStringSlice(item.Environment)
	if err != nil {
		t.Error("Wrong environment variables:", err)
	}

	if len(environment) != 2 || !contains("ONE=1", environment) || !contains("TWO=2", environment) {
		t.Error("Wrong environment variables:", item.Environment, environment)
	}

	labels, err = convert.ToStringToStringMap(item.Labels)
	if err != nil {
		t.Error("Wrong labels:", err)
	}

	if len(labels) != 2 || labels["label.one"] != "l-one" || labels["label.two"] != "l-two" {
		t.Error("Wrong labels:", item.Labels, labels)
	}

	item, err = deserialize(`
image: as_map
environment:
  ONE: '1'
  TWO: '2'
labels:
  label.one: l-one
  label.two: l-two
`)
	if err != nil {
		t.Error("Failed to deserialize:", err)
	}

	environment, err = convert.ToStringSlice(item.Environment)
	if err != nil {
		t.Error("Wrong environment variables:", err)
	}

	if len(environment) != 2 || !contains("ONE=1", environment) || !contains("TWO=2", environment) {
		t.Error("Wrong environment variables:", item.Environment, environment)
	}

	labels, err = convert.ToStringToStringMap(item.Labels)
	if err != nil {
		t.Error("Wrong labels:", err)
	}

	if len(labels) != 2 || labels["label.one"] != "l-one" || labels["label.two"] != "l-two" {
		t.Error("Wrong labels:", item.Labels, labels)
	}
}

func TestWithEnvFiles(t *testing.T) {
	f1, err := ioutil.TempFile("/tmp", "first")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("/tmp", "second")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.WriteString("ONE=1\n")
	f1.WriteString("TWO=2\n")
	f1.WriteString("OVERRIDE=12\n")
	f1.WriteString("ENV_OVERRIDE=file")
	f1.Sync()

	f2.WriteString("THREE=3\n")
	f2.WriteString("EMPTY=\n")
	f2.WriteString("# comment\n")
	f2.WriteString("OVERRIDE=42\n")
	f2.Sync()

	type EnvFiles struct {
		File1 string
		File2 string
	}

	tmpl, err := template.New("test").Parse(`
image: testing
environment:
  - STATIC=x
  - ENV_OVERRIDE=env
env_file:
  - {{.File1}}
  - {{.File2}}
`)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, EnvFiles{File1: f1.Name(), File2: f2.Name()})

	item, err := deserialize(buf.String())

	if err != nil {
		t.Error("Failed to deserialize:", err)
	}

	fromVariables, err := convert.ToStringToStringMap(item.Environment)
	if err != nil {
		t.Error("Wrong environment variables:", err)
	}

	if len(fromVariables) != 2 ||
		!mapContains("STATIC=x", fromVariables) ||
		!mapContains("ENV_OVERRIDE=env", fromVariables) {
		t.Error("Wrong environment variables:", item.Environment, fromVariables)
	}

	fromFiles, err := variablesFromEnvFiles(item.EnvFile)
	if err != nil {
		t.Error("Wrong env files:", item.EnvFile, err)
	}

	merged := mergeEnvVariables(fromFiles, fromVariables)

	if len(merged) != 7 ||
		!contains("STATIC=x", merged) ||
		!contains("ONE=1", merged) ||
		!contains("TWO=2", merged) ||
		!contains("THREE=3", merged) ||
		!contains("EMPTY=", merged) ||
		!contains("OVERRIDE=42", merged) ||
		!contains("ENV_OVERRIDE=env", merged) {
		t.Fatal("Wrong environment variables:", merged)
	}
}

func TestUlimits(t *testing.T) {
	item, err := deserialize(`
image: sample
ulimits:
  nproc: 65535
  nofile:
    soft: 20000
    hard: 40000
`)

	if err != nil {
		t.Fatal("Failed to deserialize config:", err)
	}

	ulimits, err := item.getUlimits()
	if err != nil {
		t.Fatal("Failed to parse ulimits:", err)
	}

	if len(ulimits) != 2 {
		t.Error("Invalid ulimits:", ulimits)
	}
}

func TestDefaults(t *testing.T) {
	item, err := deserialize("image: defaults")
	if err != nil {
		t.Fatal("Failed to deserialize config:", err)
	}

	if item.Image != "defaults" {
		t.Error("Wrong image")
	}

	if item.Entrypoint != nil {
		t.Error("Wrong entrypoint:", item.Entrypoint)
	}

	if item.Command != nil {
		t.Error("Wrong command:", item.Command)
	}

	if item.WorkingDir != "" {
		t.Error("Wrong working directory:", item.WorkingDir)
	}

	environment, err := convert.ToStringSlice(item.Environment)
	if err != nil {
		t.Error("Wrong environemnt variables:", err)
	}

	if len(environment) != 0 {
		t.Error("Wrong environment:", item.Environment)
	}

	labels, err := convert.ToStringToStringMap(item.Labels)
	if err != nil {
		t.Error("Wrong labels:", err)
	}

	if len(labels) != 0 {
		t.Error("Wrong labels:", item.Labels)
	}

	if item.Tty {
		t.Error("Wrong tty:", item.Tty)
	}

	if item.StopSignal != "" {
		t.Error("Wrong stop signal", item.StopSignal)
	}

	if item.StopGracePeriod != 0 {
		t.Error("Wrong stop grace period:", item.StopGracePeriod)
	}

	if item.Healthcheck != nil {
		t.Error("Wrong healthcheck:", item.Healthcheck)
	}
}

func deserialize(value string) (*Component, error) {
	var item Component

	err := yaml.UnmarshalStrict([]byte(value), &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}
