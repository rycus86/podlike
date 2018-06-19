package healthcheck

import (
	"os"
	"testing"
	"time"
)

var (
	c1  = "test-wait-1"
	id1 = "abcd0001"
	c2  = "test-wait-2"
	id2 = "abcd0002"

	results = map[string]bool{}
)

func TestState_WaitUntilStarted(t *testing.T) {
	startWait(c1, false)

	pause()
	assertNotReady(t, c1)

	MarkStarted(id1, c1)

	pause()
	assertReady(t, c1)
}

func TestState_WaitUntilHealthy(t *testing.T) {
	startWait(c2, true)

	pause()
	assertNotReady(t, c2)

	Initialize(id2, StateStarting)
	MarkStarted(id2, c2)

	pause()
	assertNotReady(t, c2)

	SetState(id2, StateHealthy)

	pause()
	assertReady(t, c2)
}

func startWait(name string, needsHealthy bool) {
	go func() {
		WaitUntilReady(name, needsHealthy)
		results[name] = true
	}()
}

func pause() {
	time.Sleep(2 * startWaitInterval)
}

func assertReady(t *testing.T, name string) {
	if ready, ok := results[name]; !ok || !ready {
		t.Error("Expected", name, "to be ready now")
	}
}

func assertNotReady(t *testing.T, name string) {
	if ready, ok := results[name]; ok || ready {
		t.Error("Not expected", name, "to be ready yet")
	}
}

func TestMain(m *testing.M) {
	originalStartWaitInterval := startWaitInterval

	startWaitInterval = 5 * time.Millisecond
	defer func() {
		startWaitInterval = originalStartWaitInterval
	}()

	os.Exit(m.Run())
}
