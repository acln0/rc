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

package rc

import "golang.org/x/sys/windows"

// CloseFunc hooks the CloseHandle system call.
var CloseFunc = windows.Close

// Init initializes the file descriptor and sets a finalizer for fd, which
// calls Close if the FD goes out of scope without being closed explicitly.
//
// If the FD was already initialized, Init returns ErrMultipleInit.
func (fd *FD) Init(sysfd windows.Handle) error {
	return fd.init(uintptr(sysfd))
}

// Incref increments the reference count associated with the FD, and returns
// the underlying file descriptor for the caller to use. The returned file
// descriptor is valid at least until the corresponding call to Decref.
//
// Callers must call Decref when they are finished using the file descriptor.
// Callers must not retain the file descriptor returned by Incref beyond
// the corresponding call to Decref.
func (fd *FD) Incref() (sysfd windows.Handle, err error) {
	uintfd, err := fd.incref()
	return windows.Handle(uintfd), err
}

func closeSysFD(sysfd uintptr) error {
	return CloseFunc(windows.Handle(sysfd))
}
