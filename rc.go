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
