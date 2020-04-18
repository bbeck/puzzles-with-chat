package main

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// SaveEnvironmentVars saves all of the environment variables and then clears
// the environment.  The saved variables are returned so that they can be
// restored later.
func SaveEnvironmentVars(t *testing.T) {
	t.Helper()

	vars := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		vars[parts[0]] = parts[1]
	}

	// When the test is complete restore the environment variables.
	t.Cleanup(func() {
		os.Clearenv()
		for key, value := range vars {
			_ = os.Setenv(key, value)
		}
	})

	os.Clearenv()
}

type CountDownLatch struct {
	sync.Mutex

	count int
	zero  chan struct{}
	done  bool
}

func NewCountDownLatch(count int) *CountDownLatch {
	return &CountDownLatch{
		count: count,
		zero:  make(chan struct{}),
		done:  count == 0,
	}
}

func (l *CountDownLatch) CountDown() {
	l.Lock()
	defer l.Unlock()

	l.count--
	if l.count <= 0 && !l.done {
		l.done = true
		close(l.zero)
	}
}

func (l *CountDownLatch) Wait(duration time.Duration) bool {
	l.Lock()

	if l.done {
		l.Unlock()
		return true
	}

	l.Unlock()

	select {
	case <-l.zero:
		return true
	case <-time.After(duration):
		return false
	}
}
