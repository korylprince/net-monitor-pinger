# About

net-monitor-pinger is used in conjunction with [Hasura GraphQL Engine](https://github.com/hasura/graphql-engine), [hasura-ad-webhook](https://github.com/korylprince/hasura-ad-webhook), and [net-monitor-ui](https://github.com/korylprince/net-monitor-ui) to form a network monitoring service.

# Install

```bash
go get github.com/korylprince/net-monitor-pinger
```

# Configuration

Hasura GraphQL Engine should be set up with the SQL schema and `metadata.json` in the [schema folder](https://github.com/korylprince/net-monitor-pinger/blob/master/schema).

The service is configured with environment variables:

```bash
DNSWorkers="8"
DNSLookupInterval="30" # in minutes
PingWorkers="16"
PingBufferSize="1024"
PingInterval="15" # in seconds
PingTimeout="1000" # in milliseconds
PurgeInterval="60" # in minutes
PurgeOlderThan="1440" # in minutes
GraphQLEndpoint="ws://example.com/v1/graphql"
GraphQLAPISecret="really long key"
```

For more information see [config.go](https://github.com/korylprince/net-monitor-pinger/blob/master/config.go).

# Docker

You can use the pre-built Docker container, [korylprince/net-monitor-pinger](https://hub.docker.com/r/korylprince/net-monitor-pinger/).

The Docker container supports [Docker Secrets](https://docs.docker.com/engine/swarm/secrets/) by appending `_FILE` to any variable, e.g. `GRAPHQLAPISECRET_FILE=/run/secrets/<secret_name>`.

## Example

```bash
docker run -d --name="net-monitor-pinger" \
    -e GRAPHQLENDPOINT="ws://example.com/v1/graphql" \
    -e GRAPHQLAPISECRET_FILE="/run/secrets/api_secret" \
    --restart="always" \
    korylprince/net-monitor-pinger:latest
```
