// Copyright 2018 Andrei Tudor Călin
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

// +build !windows

package rc

import (
	"runtime"
	"sync"

	"golang.org/x/sys/unix"
)

// CloseFunc hooks the close(2) system call.
var CloseFunc = unix.Close

// FD is a reference counted file descriptor.
//
// The zero value for FD is not usable. Values of type FD must be initialized
// by calling the Init method, and must not be copied.
//
// Once initialized, it is safe to call methods on an FD from multiple
// goroutines.
//
// Once an FD is closed, its methods return errors, and it may not be
// re-initialized.
type FD struct {
	mu          sync.RWMutex
	sysfd       int
	initialized bool
	closed      bool
}

// Init initializes the file descriptor and sets a finalizer for fd, which
// calls Close if the FD goes out of scope without being closed explicitly.
//
// If the FD was already initialized, Init returns ErrMultipleInit.
func (fd *FD) Init(sysfd int) error {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	if fd.initialized {
		return ErrMultipleInit
	}
	if fd.closed {
		return ErrClosedFD
	}
	fd.sysfd = sysfd
	fd.initialized = true
	runtime.SetFinalizer(fd, (*FD).Close)
	return nil
}

// Incref increments the reference count associated with the FD, and returns
// the underlying file descriptor for the caller to use.
//
// The returned file descriptor is valid at least until the corresponding
// call to Decref.
func (fd *FD) Incref() (sysfd int, err error) {
	fd.mu.RLock()
	if !fd.initialized {
		fd.mu.RUnlock()
		return 0, ErrUninitializedFD
	}
	if fd.closed {
		fd.mu.RUnlock()
		return 0, ErrClosedFD
	}
	return fd.sysfd, nil
}

// Decref decrements the reference count associated with the FD. Each call
// to Decref must occur after a corresponding call to Incref.
func (fd *FD) Decref() {
	fd.mu.RUnlock()
}

// Close waits for the reference count associated with the FD to reach zero,
// unsets the finalizer associated with fd, then closes the file descriptor.
func (fd *FD) Close() error {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	if !fd.initialized {
		return ErrUninitializedFD
	}
	if fd.closed {
		return ErrClosedFD
	}
	runtime.SetFinalizer(fd, nil)
	fd.closed = true
	return CloseFunc(fd.sysfd)
}
