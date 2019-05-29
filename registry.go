// Copyright 2019 Andrei Tudor Călin
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package rc

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// A LifetimeRegistry keeps track of file descriptor lifetimes for the
// purpose of testing. The zero value is ready to use.
type LifetimeRegistry struct {
	mu          sync.Mutex
	initialized int
	closed      int
	closeFailed int
	inFlight    map[int][]uintptr
}

// FDStats returns the statistics collected by the LifetimeRegistry. For
// accurate results, FDStats should be called after no more file descriptors
// are created or closed by the code under test.
func (lr *LifetimeRegistry) FDStats() FDStats {
	if lr == nil {
		return FDStats{}
	}

	lr.mu.Lock()
	defer lr.mu.Unlock()

	inFlightStacks := map[int]string{}
	for fd, pcs := range lr.inFlight {
		stack := new(strings.Builder)
		frames := runtime.CallersFrames(pcs)
		for {
			f, more := frames.Next()
			if !more {
				break
			}
			fmt.Fprintf(stack, "%s\n", f.Function)
			fmt.Fprintf(stack, "\t%s:%d\n", f.File, f.Line)
		}
		inFlightStacks[fd] = stack.String()
	}

	return FDStats{
		Initialized:    lr.initialized,
		Closed:         lr.closed,
		CloseFailed:    lr.closeFailed,
		InFlightStacks: inFlightStacks,
	}
}

// FDStats is a set of file descriptor statistics.
type FDStats struct {
	// Initialized is the number of initialized file descriptors.
	Initialized int

	// Closed is the number of closed file descriptors.
	Closed int

	// CloseFailed is the number of file descriptors for which
	// the Close method failed.
	CloseFailed int

	// InFlightStacks maps file descriptor numbers to goroutine stack
	// traces taken their initialization sites.
	InFlightStacks map[int]string
}

// Report returns a report of file descriptor stats. If no file descriptors
// were leaked, then Report returns the empty string. Otherwise, it returns
// a message suitable for usage when failing a test.
func (stats FDStats) Report() string {
	if !stats.leakedFDs() {
		return ""
	}

	report := new(strings.Builder)
	fmt.Fprint(report, "file descriptor report:\n")
	fmt.Fprintf(report, "initialized %d FDs\n", stats.Initialized)
	fmt.Fprintf(report, "closed %d FDs successfully\n", stats.Closed)
	fmt.Fprintf(report, "closed %d FDs unsuccessfully\n", stats.CloseFailed)
	fmt.Fprint(report, "file descriptors in flight:\n")
	fmt.Fprint(report, "----------------\n")
	for fd, stack := range stats.InFlightStacks {
		fmt.Fprintf(report, "FD=%d initialized at:\n", fd)
		fmt.Fprintf(report, stack)
		fmt.Fprint(report, "----------------\n")
	}
	return report.String()
}

// leakedFDs returns a boolean indicating whether there exist file descriptors
// for which Close was never called.
func (stats FDStats) leakedFDs() bool {
	closed := stats.Closed + stats.CloseFailed
	return stats.Initialized != closed
}

// recordInit records an Init call for the specified file descriptor.
func (lr *LifetimeRegistry) recordInit(fd int) {
	if lr == nil {
		return
	}

	lr.mu.Lock()
	defer lr.mu.Unlock()

	lr.initialized++
	if lr.inFlight == nil {
		lr.inFlight = map[int][]uintptr{}
	}
	lr.inFlight[fd] = lr.callers()
}

func (lr *LifetimeRegistry) recordClose(fd int, err error) {
	if lr == nil {
		return
	}

	lr.mu.Lock()
	defer lr.mu.Unlock()

	if err != nil {
		lr.closeFailed++
	} else {
		lr.closed++
	}

	delete(lr.inFlight, fd)
}

func (lr *LifetimeRegistry) callers() []uintptr {
	// this function, recordInit, (*FD.Init)
	const skip = 4

	pcs := make([]uintptr, 50)
	n := runtime.Callers(skip, pcs)
	return pcs[:n]
}
