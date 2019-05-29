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

// Package rc provides reference counted file descriptors.
//
// FD is a low level construct, and is useful only under very specific
// circumstances. In most use cases, managing file descriptors using
// the standard library os or net packages is a better choice.
package rc

import (
	"errors"
	"os"
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
	// Init was already called on FD.
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
// FD is not suitable for use with blocking system calls: its internal
// locking scheme assumes that calls to Do do not block for very long.
//
// Once an FD is closed, its methods return errors, and it may not be
// re-initialized.
type FD struct {
	lr *LifetimeRegistry // ok to be nil

	mu          sync.RWMutex
	rawfd       int
	initialized bool
	closed      bool
	closeFunc   func(int) error
}

// TrackLifetime instructs the FD to reports its lifetime cycle to the specified
// LifetimeRegistry. For accurate results, TrackLifetime must be called before
// Init.
func (fd *FD) TrackLifetime(lr *LifetimeRegistry) {
	fd.lr = lr
}

// Init initializes the file descriptor and sets a finalizer for fd, which may
// call closeFunc if the FD goes out of scope without being closed explicitly.
//
// If the FD was already initialized, Init returns ErrMultipleInit.
func (fd *FD) Init(rawfd int, closeFunc func(int) error) error {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	if fd.closed {
		return ErrClosedFD
	}
	if fd.initialized {
		return ErrMultipleInit
	}
	fd.rawfd = rawfd
	fd.initialized = true
	fd.closeFunc = closeFunc
	fd.lr.recordInit(rawfd)
	runtime.SetFinalizer(fd, (*FD).Close)
	return nil
}

// Do executes fn against the file descriptor. If Do does not return an
// error, the file descriptor is guaranteed to be valid for the duration of
// the call to fn.
func (fd *FD) Do(fn func(rawfd int) error) error {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	if !fd.initialized {
		return ErrUninitializedFD
	}
	if fd.closed {
		return ErrClosedFD
	}
	return fn(fd.rawfd)
}

// Close waits for the reference count associated with the FD to reach zero,
// unsets the finalizer associated with fd, then closes the file descriptor.
//
// Calling Close from inside a Do block causes a deadlock, so it is forbidden.
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
	err := fd.closeFunc(fd.rawfd)
	fd.lr.recordClose(fd.rawfd, err)
	return err
}

// WrapSyscallError wraps an error from a call to (*FD).Do or (*FD).Close,
// with a few special cases taken into consideration:
//
// The sentinel error values ErrUninitializedFD and ErrClosedFD are returned
// without wrapping. If err is nil, WrapSyscallError returns nil. Any other
// error is wrapped using os.NewSyscallError.
func WrapSyscallError(syscall string, err error) error {
	switch err {
	case nil:
		return nil
	case ErrUninitializedFD, ErrClosedFD:
		return err
	default:
		return os.NewSyscallError(syscall, err)
	}
}
