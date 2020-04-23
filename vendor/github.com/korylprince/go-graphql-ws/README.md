[![pkg.go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/korylprince/go-graphql-ws)

# About

`go-graphql-ws` is a Go client implementation of the [GraphQL over WebSocket Protocol](https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md). This is not a standard GraphQL client; the GraphQL endpoint must support the WebSocket Protocol. This package has been written specifically with [Hasura GraphQL Engine](https://github.com/hasura/graphql-engine) in mind, but any server implementing the WebSocket Protocol should work.

## Package Status

This package is still under testing; use it with caution. A `v1` version has not been released yet so the API should not be considered stable.

# Installing

`go get github.com/korylprince/go-graphql-ws`

# Usage

See [Examples](https://pkg.go.dev/github.com/korylprince/go-graphql-ws?tab=doc#pkg-examples).

# Issues

If you have any issues or questions [create an issue](https://github.com/korylprince/go-graphql-ws/issues).
