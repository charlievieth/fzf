package util

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var atExitFuncs struct {
	fns []func()
	sync.Mutex
}

func AtExit(fn func()) {
	atExitFuncs.Lock()
	defer atExitFuncs.Unlock()
	atExitFuncs.fns = append(atExitFuncs.fns, fn)
}

func Exit(code int) {
	atExitFuncs.Lock()
	defer atExitFuncs.Unlock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, fn := range atExitFuncs.fns {
			if fn != nil {
				fn()
			}
		}
	}()
	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 5):
		fmt.Fprintln(os.Stderr, "Error: timed out waiting for atexit funcs")
	}
	os.Exit(code)
}
