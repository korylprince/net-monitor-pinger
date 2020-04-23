package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

//GenerateSubscriptionID is a function that returns unique IDs used to track subscriptions.
//By default UUIDv4's are used
var GenerateSubscriptionID func() string = func() string {
	return uuid.Must(uuid.NewV4()).String()
}

//Conn is a connection to a GraphQL WebSocket endpoint
type Conn struct {
	conn  *websocket.Conn
	debug bool

	subscriptions map[string]func(message *Message)
	mu            *sync.RWMutex

	closeError   error
	closeHandler func(code int, text string)
}

func (c *Conn) reader() {
	for {
		msg := new(Message)
		err := c.conn.ReadJSON(msg)
		if websocket.IsUnexpectedCloseError(err) {
			c.closeError = err
			if c.closeHandler != nil {
				cErr := err.(*websocket.CloseError)
				c.closeHandler(cErr.Code, cErr.Text)
			}
			if c.debug {
				log.Println("DEBUG: Connection closed:", err)
			}
			return
		}
		if c.debug && err != nil {
			log.Println("DEBUG: Unable to parse Message:", err)
			continue
		}

		if msg.Type == MessageTypeConnectionKeepAlive {
			continue
		}

		c.mu.RLock()
		if f, ok := c.subscriptions[msg.ID]; !ok {
			if c.debug {
				fmt.Printf("DEBUG: Message received for unknown subscription: ID: %s, Type: %s, Payload: %s\n", msg.ID, msg.Type, string(msg.Payload))
			}
		} else {
			go f(msg)
		}
		c.mu.RUnlock()

		if msg.Type == MessageTypeComplete && msg.ID != "" {
			c.mu.Lock()
			delete(c.subscriptions, msg.ID)
			c.mu.Unlock()
		}

		if msg.Type != MessageTypeComplete && msg.Type != MessageTypeData && msg.Type != MessageTypeError && c.debug {
			fmt.Printf("DEBUG: Received Message with unexpected type: ID: %s, Type: %s, Payload: %s\n", msg.ID, msg.Type, string(msg.Payload))
		}
	}
}

func (c *Conn) init(connectionParams *MessagePayloadConnectionInit) error {
	msg := &Message{Type: MessageTypeConnectionInit}
	if err := msg.SetPayload(connectionParams); err != nil {
		return fmt.Errorf("Unable to marshal connectionParams: %v", err)
	}

	err := c.conn.WriteJSON(msg)
	if err != nil {
		return fmt.Errorf("Unable to write %s message: %v", MessageTypeConnectionInit, err)
	}

	for {
		msg := new(Message)
		err = c.conn.ReadJSON(msg)
		if websocket.IsUnexpectedCloseError(err) {
			return fmt.Errorf("Unexpected close error: %v", err)
		}
		if err != nil {
			return fmt.Errorf("Unable to parse message: %v", err)
		}
		switch msg.Type {
		case MessageTypeConnectionAck:
			return nil
		case MessageTypeConnectionKeepAlive:
			continue
		case MessageTypeConnectionError:
			return ParseError(msg.Payload)
		default:
			return fmt.Errorf("Unexpected message type: %s", msg.Type)
		}
	}
}

//Close closes the Conn or returns an error if one occurred
func (c *Conn) Close() error {
	if c.closeError != nil {
		return c.closeError
	}

	err := c.conn.WriteJSON(&Message{Type: MessageTypeConnectionTerminate})
	if err != nil {
		return fmt.Errorf("Unable to write %s message: %v", MessageTypeConnectionTerminate, err)
	}

	err = c.conn.Close()
	if err != nil {
		return fmt.Errorf("Unable to close websocket connection: %v", err)
	}

	return nil
}

//SetCloseHandler sets the handler for when the Conn is closed, expectedly or not.
//The code argument to h is the received close code or CloseNoStatusReceived if the close message is empty
func (c *Conn) SetCloseHandler(f func(code int, text string)) {
	c.closeHandler = f
}

//Subscribe creates a GraphQL subscription with the given payload and returns its ID, or returns an error if one occurred.
//Subscription Messages are passed to the given function handler as they are received
func (c *Conn) Subscribe(payload *MessagePayloadStart, f func(message *Message)) (id string, err error) {
	if c.closeError != nil {
		return "", c.closeError
	}

	id = GenerateSubscriptionID()

	m := &Message{Type: MessageTypeStart, ID: id}
	if err := m.SetPayload(payload); err != nil {
		return "", fmt.Errorf("Unable to marshal payload: %v", err)
	}

	c.mu.Lock()
	c.subscriptions[id] = f
	c.mu.Unlock()

	if err := c.conn.WriteJSON(m); err != nil {
		c.mu.Lock()
		delete(c.subscriptions, id)
		c.mu.Unlock()
		return "", fmt.Errorf("Unable to write %s message: %v", MessageTypeStart, err)
	}

	return id, nil
}

//Unsubscribe stops the subscription with the given ID or returns an error if one occurred
func (c *Conn) Unsubscribe(id string) error {
	if c.closeError != nil {
		return c.closeError
	}

	m := &Message{Type: MessageTypeStop, ID: id}

	if err := c.conn.WriteJSON(m); err != nil {
		return fmt.Errorf("Unable to write %s message: %v", MessageTypeStop, err)
	}

	//if subscription still exists wait up to 5 seconds for complete message to arrive before deleting it
	//this only matters in debug mode since the message is ignored when debug mode is off
	if c.debug {
		c.mu.Lock()
		if _, ok := c.subscriptions[id]; ok {
			c.subscriptions[id] = func(message *Message) {}
			go func() {
				time.Sleep(time.Second * 5)
				c.mu.Lock()
				delete(c.subscriptions, id)
				c.mu.Unlock()
			}()
		}
		c.mu.Unlock()
	} else {
		c.mu.Lock()
		delete(c.subscriptions, id)
		c.mu.Unlock()
	}

	return nil
}

//Execute executes the given payload and returns the result or an error if one occurred
//The given context can be used to cancel the request
func (c *Conn) Execute(ctx context.Context, payload *MessagePayloadStart) (data *MessagePayloadData, err error) {
	if c.closeError != nil {
		return nil, c.closeError
	}

	ch := make(chan *Message)
	id, err := c.Subscribe(payload, func(message *Message) {
		ch <- message
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to subscribe: %v", err)
	}

	defer func() {
		if uErr := c.Unsubscribe(id); err == nil && uErr != nil {
			err = fmt.Errorf("Unable to unsubscribe: %v", uErr)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg := <-ch:
			switch msg.Type {
			case MessageTypeComplete:
				continue
			case MessageTypeData:
				d := new(MessagePayloadData)
				if err = json.Unmarshal(msg.Payload, d); err != nil {
					return nil, fmt.Errorf("Unable to unmarshal %s message payload: %v", MessageTypeData, err)
				}
				return d, nil
			case MessageTypeError:
				return nil, ParseError(msg.Payload)
			default:
				return nil, fmt.Errorf("Unexpected message type: %s", msg.Type)
			}
		}
	}
}
