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

package rc_test

import (
	"errors"
	"testing"

	"acln.ro/rc/v2"
)

func TestFD(t *testing.T) {
	t.Run("BasicInit", testBasicInit)
	t.Run("MultipleInit", testMultipleInit)
	t.Run("InitClosed", testInitClosed)
	t.Run("DoUninitialized", testDoUninitialized)
	t.Run("DoClosed", testDoClosed)
	t.Run("DoReturnsInnerError", testDoReturnsInnerError)
	t.Run("CloseUninitialized", testCloseUninitialized)
	t.Run("DoubleClose", testDoubleClose)
}

func testBasicInit(t *testing.T) {
	fd := new(rc.FD)
	want := 42
	if err := fd.Init(42, dummyClose); err != nil {
		t.Fatalf("Init: %v", err)
	}
	var got int
	readRawFD := func(rawfd int) error {
		got = rawfd
		return nil
	}
	if err := fd.Do(readRawFD); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got != want {
		t.Fatalf("Do after Init: got %d, want %d", got, want)
	}
}

func testMultipleInit(t *testing.T) {
	fd := new(rc.FD)
	if err := fd.Init(42, dummyClose); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	switch err := fd.Init(43, dummyClose); err {
	case nil:
		t.Fatal("second Init: did not fail")
	case rc.ErrMultipleInit:
		// ok
	default:
		t.Fatalf("second Init: got %v, want ErrMultipleInit", err)
	}
}

func testInitClosed(t *testing.T) {
	fd := new(rc.FD)
	if err := fd.Init(42, dummyClose); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := fd.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	switch err := fd.Init(42, dummyClose); err {
	case nil:
		t.Fatal("Init after Close succeeded")
	case rc.ErrClosedFD:
		// ok
	default:
		t.Fatalf("Init after Close: got %v, want ErrClosedFD", err)
	}
}

func testDoUninitialized(t *testing.T) {
	fd := new(rc.FD)
	switch err := fd.Do(dummyDo); err {
	case nil:
		t.Fatal("Do succeded on uninitialized FD")
	case rc.ErrUninitializedFD:
		// ok
	default:
		t.Fatalf("Do on uninitialized FD: got %v, want ErrUninitializedFD", err)
	}
}

func testDoClosed(t *testing.T) {
	fd := new(rc.FD)
	if err := fd.Init(42, dummyClose); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := fd.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	switch err := fd.Do(dummyDo); err {
	case nil:
		t.Fatal("Do succeeded on closed FD")
	case rc.ErrClosedFD:
		// ok
	default:
		t.Fatalf("Do on closed FD: got %v, want ErrClosedFD", err)
	}
}

func testDoReturnsInnerError(t *testing.T) {
	fd := new(rc.FD)
	if err := fd.Init(42, dummyClose); err != nil {
		t.Fatalf("Init: %v", err)
	}
	testError := errors.New("test error")
	fn := func(_ int) error {
		return testError
	}
	if err := fd.Do(fn); err != testError {
		t.Fatalf("Do: want %v, got %v", testError, err)
	}
}

func testCloseUninitialized(t *testing.T) {
	fd := new(rc.FD)
	switch err := fd.Close(); err {
	case nil:
		t.Fatal("Close succeeded on uninitialized FD")
	case rc.ErrUninitializedFD:
		// ok
	default:
		t.Fatalf("Close on uninitialized FD: got %v, want ErrUninitializedFD", err)
	}
}

func testDoubleClose(t *testing.T) {
	closeCalls := 0
	countingClose := func(_ int) error {
		closeCalls++
		return nil
	}
	fd := new(rc.FD)
	if err := fd.Init(42, countingClose); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := fd.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if closeCalls != 1 {
		t.Fatalf("Close did not call closeFunc")
	}
	switch err := fd.Close(); err {
	case nil:
		t.Fatalf("Close succeeded on closed FD; %d calls to closeFunc", closeCalls)
	case rc.ErrClosedFD:
		if closeCalls != 1 {
			t.Fatalf("Close called closeFunc too many times: got %d, want %d", closeCalls, 1)
		}
	default:
		t.Fatalf("Close on closed FD: got %v, want ErrClosedFD", err)
	}
}

func dummyDo(_ int) error { return nil }

func dummyClose(_ int) error { return nil }
