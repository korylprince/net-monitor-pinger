[![pkg.go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/korylprince/go-icmpv4/v2)

# About

`go-icmpv4` is a library for working with ICMPv4 packets. The `echo` helper library is useful for dealing with ICMPv4 Echo Request/Reply (ping) packets.

# Installing

Using Go Modules:

`go get github.com/korylprince/go-icmpv4/v2`

Using gopkg.in:

`go get gopkg.in/korylprince/go-icmpv4.v2`

# Usage

See the [package echo docs](https://pkg.go.dev/github.com/korylprince/go-icmpv4/v2/echo?tab=doc) for example usage.

# Issues

If you have any issues or questions [create an issue](https://github.com/korylprince/go-icmpv4/issues).

# Testing

`go test -v`

Note: testing will need to be run with raw socket privileges (i.e. with `sudo`.)
