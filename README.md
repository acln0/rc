acln.ro/rc
================

[![GoDoc](https://godoc.org/acln.ro/rc?status.svg)](https://godoc.org/acln.ro/rc)

Package rc provides reference-counted file descriptors.

This package solves a very niche problem, namely managing the lifetime
of a file descriptor which is not packaged and used as an `*os.File`,
or as a `net.Conn` / `net.PacketConn`, in the presence of potential
concurrent access.

I originally built package rc to help manage eBPF map file descriptors,
but the code is generic enough that it can be used for other similar
purposes just as well.

### Package version

Package rc presents itself as a Go module, and is currently at v2.0.0.
The update from v1 to v2 changed the API to be function-based, and
removed special handling of Windows file descriptors.

### License

Package rc is distributed under the ISC license. A copy of the license
can be found in the LICENSE file.
