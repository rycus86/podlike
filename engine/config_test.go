package engine

import (
	"gopkg.in/yaml.v2"
	"testing"
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

	if len(item.Environment) != 1 || item.Environment[0] != "ENV=environment" {
		t.Error("Wrong environment variables:", item.Environment)
	}

	if len(item.Labels) != 2 {
		t.Error("Wrong number of labels:", item.Labels)
	}

	if value, ok := item.Labels["label.one"]; !ok || value != "first label" {
		t.Error("Wrong labels:", item.Labels)
	}

	if value, ok := item.Labels["label.two"]; !ok || value != "second label" {
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

	if len(item.Environment) != 0 {
		t.Error("Wrong environment:", item.Environment)
	}

	if len(item.Labels) != 0 {
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
