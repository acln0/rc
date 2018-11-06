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
// FD is a low level construct, and is useful only under very specific,
// rare circumstances. More often than not, callers should use the standard
// library os package to manage raw file descriptors instead.
package rc

import "errors"

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
