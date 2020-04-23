//Package graphql implements a client for the GraphQL over WebSocket Protocol described at https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md.
package graphql

import "encoding/json"

const (
	MessageTypeConnectionInit      = "connection_init"
	MessageTypeConnectionAck       = "connection_ack"
	MessageTypeConnectionError     = "connection_error"
	MessageTypeConnectionTerminate = "connection_terminate"
	MessageTypeConnectionKeepAlive = "ka"
	MessageTypeStart               = "start"
	MessageTypeData                = "data"
	MessageTypeComplete            = "complete"
	MessageTypeStop                = "stop"
	MessageTypeError               = "error"
)

type MessagePayloadConnectionInit map[string]interface{}
type MessagePayloadStart struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}
type MessagePayloadData struct {
	Data   json.RawMessage `json:"data"`
	Errors []Error         `json:"errors,omitempty"`
}

//Message is a GraphQL message. A Message's Payload can be JSON decoded into one of the MessagePayload* types if present
type Message struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

//SetPayload sets Message.Payload to the given payload or returns an error if one occurred
func (m *Message) SetPayload(payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	m.Payload = json.RawMessage(b)
	return nil
}
