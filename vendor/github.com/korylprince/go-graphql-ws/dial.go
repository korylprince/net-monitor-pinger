package graphql

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

//Dialer is a wrapper for websocket.Dialer
type Dialer struct {
	*websocket.Dialer
	Debug bool
}

//DefaultDialer is a wrapper for websocket.DeafultDialer
var DefaultDialer = &Dialer{Dialer: websocket.DefaultDialer}

//Dial creates a new Conn by calling DialContext with a background context
func (d *Dialer) Dial(urlStr string, requestHeader http.Header, connectionParams *MessagePayloadConnectionInit) (*Conn, *http.Response, error) {
	return d.DialContext(context.Background(), urlStr, requestHeader, connectionParams)
}

//DialContext creates a new Conn with the same parameters as websocket.DialContext.
//connectionParams is passed to the GraphQL server and is described further at
//https://www.apollographql.com/docs/react/data/subscriptions/#authentication-over-websocket
func (d *Dialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header, connectionParams *MessagePayloadConnectionInit) (*Conn, *http.Response, error) {
	conn, resp, err := d.Dialer.DialContext(ctx, urlStr, requestHeader)
	if err != nil {
		return nil, resp, fmt.Errorf("Unable to dial: %v", err)
	}

	c := &Conn{
		conn:          conn,
		debug:         d.Debug,
		subscriptions: make(map[string]func(message *Message)),
		mu:            new(sync.RWMutex),
	}

	if err = c.init(connectionParams); err != nil {
		return nil, resp, fmt.Errorf("Unable to init: %v", err)
	}

	go c.reader()

	return c, resp, nil
}
