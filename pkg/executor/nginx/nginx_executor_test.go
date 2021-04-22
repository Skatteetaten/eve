package nginx

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

var settleTime = 1000 * time.Millisecond

func TestRotateSignal(t *testing.T) {

	file, err := ioutil.TempFile("/tmp", "logrotate")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	r := nginxLogRotate{
		paths:           []string{file.Name()},
		rotateAfterSize: 0,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	pid := syscall.Getpid()
	logrus.Infof("Pid: %d", pid)

	if err := r.rotate(pid, file.Name()); err != nil {
		t.Fatalf("Log rotate failed")
	}

	waitSig(t, c, syscall.SIGUSR1)
}

func TestHandleExit(t *testing.T) {
	executor := NewNginxExecutor(0, nil)

	termCode := int(syscall.SIGTERM) + 128
	termHandled := executor.HandleExit(termCode, 1)
	if termHandled != 0 {
		t.Errorf("Handler returned wrong value. Got %d, want %d", termHandled, 0)
	}

	sigintCode := int(syscall.SIGINT) + 128
	sigIntHandled := executor.HandleExit(sigintCode, 1)
	if sigIntHandled != 0 {
		t.Errorf("Handler returned wrong value. Got %d, want %d", sigIntHandled, 0)
	}
}

func TestHandleLogRotate(t *testing.T) {

	file, err := ioutil.TempFile("/tmp", "logrotate")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	e := NewNginxExecutor(0, []string{file.Name()})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	pid := syscall.Getpid()
	e.StartLogRotate(pid, 600)

	waitSig(t, c, syscall.SIGUSR1)
}

func TestPrepareCommand(t *testing.T) {
	e := NewNginxExecutor(0, nil)
	cmd := e.PrepareForNginxRun("/tmp/nginx/nginx.conf")
	assert.Equal(t, "sh", cmd.Args[0])
	assert.Equal(t, "-c", cmd.Args[1])
	assert.Equal(t, "exec nginx -g 'daemon off;' -c /tmp/nginx/nginx.conf", cmd.Args[2])

}

//https://golang.org/src/os/signal/signal_test.go
func waitSig(t *testing.T, c <-chan os.Signal, sig os.Signal) {

	t.Helper()

	waitSig1(t, c, sig, false)
}

//https://golang.org/src/os/signal/signal_test.go
func waitSig1(t *testing.T, c <-chan os.Signal, sig os.Signal, all bool) {

	t.Helper()

	// Sleep multiple times to give the kernel more tries to

	// deliver the signal.

	start := time.Now()

	timer := time.NewTimer(settleTime / 10)

	defer timer.Stop()

	// If the caller notified for all signals on c, filter out SIGURG,

	// which is used for runtime preemption and can come at unpredictable times.

	// General user code should filter out all unexpected signals instead of just

	// SIGURG, but since os/signal is tightly coupled to the runtime it seems

	// appropriate to be stricter here.

	for time.Since(start) < settleTime {

		select {

		case s := <-c:

			if s == sig {

				return

			}

			if !all || s != syscall.SIGURG {

				t.Fatalf("signal was %v, want %v", s, sig)

			}

		case <-timer.C:

			timer.Reset(settleTime / 10)

		}

	}

	t.Fatalf("timeout after %v waiting for %v", settleTime, sig)

}