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
	"reflect"
	"strings"
	"testing"

	"acln.ro/rc/v2"
)

func TestLifetimeRegistry(t *testing.T) {
	t.Run("RecordInit", testLifetimeRegistryRecordInit)
	t.Run("RecordClose", testLifetimeRegistryRecordClose)
	t.Run("Report", testLifetimeRegistryReport)
	t.Run("NilSafety", testLifetimeRegistryNilSafety)
}

func testLifetimeRegistryRecordInit(t *testing.T) {
	var lreg rc.LifetimeRegistry
	fd := rc.FD{}

	fd.TrackLifetime(&lreg)
	if err := fd.Init(42, dummyClose); err != nil {
		t.Fatal(err)
	}

	if stats := lreg.FDStats(); stats.Initialized != 1 {
		t.Fatalf("didn't register initialized FD")
	}
}

func testLifetimeRegistryRecordClose(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var lreg rc.LifetimeRegistry

		fd := rc.FD{}
		fd.TrackLifetime(&lreg)
		fd.Init(42, dummyClose)
		fd.Close()

		if stats := lreg.FDStats(); stats.Closed != 1 {
			t.Fatalf("didn't record close")
		}
	})
	t.Run("Error", func(t *testing.T) {
		var lreg rc.LifetimeRegistry

		fd := rc.FD{}
		fd.TrackLifetime(&lreg)
		fd.Init(42, func(_ int) error { return errors.New("bad") })
		fd.Close()

		if stats := lreg.FDStats(); stats.CloseFailed != 1 {
			t.Fatalf("didn't record failed close")
		}
	})
}

func testLifetimeRegistryReport(t *testing.T) {
	t.Run("NoLeak", testLifetimeRegistryReportNoLeak)
	t.Run("Leak", testLifetimeRegistryReportLeak)
}

func testLifetimeRegistryReportNoLeak(t *testing.T) {
	var lreg rc.LifetimeRegistry

	var fd1, fd2 rc.FD

	fd1.TrackLifetime(&lreg)
	fd2.TrackLifetime(&lreg)

	if err := fd1.Init(1, dummyClose); err != nil {
		t.Fatal(err)
	}
	if err := fd2.Init(2, dummyClose); err != nil {
		t.Fatal(err)
	}

	if err := fd1.Close(); err != nil {
		t.Fatal(err)
	}
	if err := fd2.Close(); err != nil {
		t.Fatal(err)
	}

	stats := lreg.FDStats()
	if stats.Initialized != 2 {
		t.Errorf("got Initialized == %d, want 2", stats.Initialized)
	}
	if stats.Closed != 2 {
		t.Errorf("got Closed == %d, want 2", stats.Closed)
	}
	if stats.CloseFailed != 0 {
		t.Errorf("got CloseFailed == %d, want 0", stats.CloseFailed)
	}
	if report := stats.Report(); report != "" {
		t.Errorf("got non-empty report %s", report)
	}
}

func testLifetimeRegistryReportLeak(t *testing.T) {
	var lreg rc.LifetimeRegistry

	var fd1, fd2, fd3 rc.FD

	fd1.TrackLifetime(&lreg)
	fd2.TrackLifetime(&lreg)
	fd3.TrackLifetime(&lreg)

	functionOne(&fd1)
	functionTwo(&fd2)
	functionThree(&fd3)

	stats := lreg.FDStats()
	if stats.Initialized != 3 {
		t.Errorf("got Initialized == %d, want 3", stats.Initialized)
	}
	if stats.Closed != 1 {
		t.Errorf("got Closed == %d, want 1", stats.Closed)
	}
	report := stats.Report()
	if report == "" {
		t.Errorf("got empty report")
	}

	t.Log(report)

	if !strings.Contains(report, "FD=43") {
		t.Errorf("missing FD=43 from report")
	}
	if !strings.Contains(report, "FD=44") {
		t.Errorf("missing FD=44 from report")
	}
	if !strings.Contains(report, "functionTwo") {
		t.Errorf("missing functionTwo from report")
	}
	if !strings.Contains(report, "functionThree") {
		t.Errorf("missing functionThree from report")
	}
}

func functionOne(fd *rc.FD) {
	fd.Init(42, dummyClose)
	fd.Close()
}

func functionTwo(fd *rc.FD) {
	fd.Init(43, dummyClose)
}

func functionThree(fd *rc.FD) {
	fd.Init(44, dummyClose)
}

func testLifetimeRegistryNilSafety(t *testing.T) {
	var lreg *rc.LifetimeRegistry

	fd := rc.FD{}
	fd.TrackLifetime(lreg)
	fd.Init(42, dummyClose)
	fd.Close()

	stats := lreg.FDStats()
	if !reflect.DeepEqual(stats, rc.FDStats{}) {
		t.Errorf("got %#v, want empty stats", stats)
	}
}
