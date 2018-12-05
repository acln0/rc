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

// Package rc provides reference counted file descriptors.
//
// FD is a low level construct, and is useful only under very specific
// circumstances. In most use cases, managing file descriptors using
// the standard library os or net packages is a better choice.
package rc

import (
	"errors"
	"runtime"
	"sync"
)

var (
	// ErrUninitializedFD is the error returned by FD methods when called
	// on a file descriptor which has not been initialized.
	ErrUninitializedFD = errors.New("rc: use of uninitialized file descriptor")

	// ErrClosedFD is the error returned by FD methods when called on
	// a file descriptor which has been closed.
	ErrClosedFD = errors.New("rc: use of closed file descriptor")

	// ErrMultipleInit is the error returned by (*FD).Init when called
	// for at least the second time on a specific FD.
	ErrMultipleInit = errors.New("rc: multiple calls to (*FD).Init")
)

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
	sysfd       uintptr
	initialized bool
	closed      bool
}

func (fd *FD) init(sysfd uintptr) error {
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

func (fd *FD) incref() (uintptr, error) {
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

// Decref decrements the reference count associated with the FD. Calls to
// to Decref must occur after corresponding calls to Incref.
func (fd *FD) Decref() {
	fd.mu.RUnlock()
}

// Close waits for the reference count associated with the FD to reach zero,
// unsets the finalizer associated with fd, then closes the file descriptor.
//
// Close cannot be called while holding a reference to the FD (i.e. between
// an Incref and a Decref).
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
	return closeSysFD(fd.sysfd)
}
