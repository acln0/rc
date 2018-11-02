// Copyright 2018 Andrei Tudor Călin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rc

import (
	"runtime"
	"sync"

	"golang.org/x/sys/windows"
)

// CloseFunc hooks the CloseHandle system call.
var CloseFunc = windows.Close

// FD is a reference counted file descriptor.
//
// The zero value for FD is not usable. Values of type FD must be initialized
// by calling the Init method, and must not be copied.
//
// Once initialized, it is safe to call methods on an FD from multiple
// goroutines.
//
// Once an FD is closed, it may not be re-initialized.
type FD struct {
	mu          sync.RWMutex
	sysfd       windows.Handle
	initialized bool
	closed      bool
}

// Init initializes the file descriptor and sets a finalizer for fd, which
// calls Close if the FD goes out of scope without being closed explicitly.
//
// If the FD was already initialized, Init returns ErrMultipleInit.
func (fd *FD) Init(sysfd windows.Handle) error {
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
func (fd *FD) Incref() (sysfd windows.Handle, err error) {
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
