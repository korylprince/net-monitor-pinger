FROM golang:1.14-alpine as builder

ARG VERSION

RUN apk add --no-cache git

RUN GO111MODULE=on go get github.com/korylprince/fileenv@v1.1.0

RUN git clone --branch "$VERSION" --single-branch --depth 1 \
    https://github.com/korylprince/net-monitor-pinger.git  /go/src/github.com/korylprince/net-monitor-pinger

RUN cd /go/src/github.com/korylprince/net-monitor-pinger && \
    go install -mod=vendor github.com/korylprince/net-monitor-pinger


FROM alpine:3.11

RUN apk add --no-cache bash ca-certificates

COPY --from=builder /go/bin/fileenv /
COPY --from=builder /go/bin/net-monitor-pinger /

CMD ["/fileenv", "/net-monitor-pinger"]
