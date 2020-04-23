package graphql

import (
	"encoding/json"
)

//Location is a location in a GraphQL document
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

//Error is a GraphQL Error
type Error struct {
	Message    string                 `json:"message"`
	Locations  []Location             `json:"locations"`
	Path       []interface{}          `json:"path"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

//Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

type Errors []Error

//Error implements the error interface
func (e Errors) Error() string {
	if len(e) > 0 {
		return e[0].Error()
	}
	return "no error"
}

type UnknownError struct {
	Original []byte
}

//Error implements the error interface
func (e *UnknownError) Error() string {
	return string(e.Original)
}

//ParseError parses and returns an error for the various different formats a GraphQL server might return an error
func ParseError(payload []byte) error {
	type wrapper struct {
		Errors Errors `json:"errors"`
	}
	w := new(wrapper)
	if err := json.Unmarshal(payload, w); err == nil && w.Errors != nil {
		return w.Errors
	}

	es := make(Errors, 0)
	if err := json.Unmarshal(payload, &es); err == nil {
		return es
	}

	e := new(Error)
	if err := json.Unmarshal(payload, e); err == nil && e.Message != "" {
		return e
	}

	return &UnknownError{Original: payload}
}
