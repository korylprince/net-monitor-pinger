package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/korylprince/go-graphql-ws"
)

const gqlSubscribeDevices = `
	subscription get_devices {
	  device {
		id
		hostname
	  }
	}
`

const gqlInsertPings = `
	mutation insert_ping($pings: [ping_insert_input!]!) {
	  insert_ping(objects: $pings) {
		affected_rows
	  }
	}
`

const gqlPurgePings = `
	mutation purge_pings($time: timestamp!) {
	  delete_ping(where: {sent_time: {_lt: $time}}) {
		affected_rows
	  }
	}
`

type GraphQLService struct {
	conn             *graphql.Conn
	subscribeHandler func(devices []*Device)
}

func NewGraphQLService(endpoint string, apiKey string) (*GraphQLService, error) {
	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	headers.Set("X-Authorization-Type", "API-Key")
	conn, _, err := graphql.DefaultDialer.Dial(endpoint, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect: %v", err)
	}

	g := &GraphQLService{conn: conn}
	conn.SetCloseHandler(func(code int, text string) {
		log.Println("GraphQLService: WebSocket closed:", text)
		g.reconnect(endpoint, headers)
	})

	log.Println("GraphQLService: Connected to:", endpoint)

	return g, nil
}

func (g *GraphQLService) reconnect(endpoint string, headers http.Header) {
	retry := time.Second
	for {
		log.Println("GraphQLService: Retrying connection in", retry.String())
		time.Sleep(retry)
		retry *= 2

		conn, _, err := graphql.DefaultDialer.Dial(endpoint, headers, nil)
		if err != nil {
			log.Println("GraphQLService: Unable to reconnect:", err)
			continue
		}

		g.conn = conn
		if g.subscribeHandler != nil {
			if err = g.subscribeDevices(); err != nil {
				log.Println("GraphQLService: Unable to resubscribe:", err)
				continue
			}
		}

		conn.SetCloseHandler(func(code int, text string) {
			log.Println("GraphQLService: WebSocket closed:", text)
			g.reconnect(endpoint, headers)
		})
		return
	}
}

func (g *GraphQLService) subscribeDevices() error {
	type response struct {
		Devices []*Device `json:"device"`
	}

	var q = &graphql.MessagePayloadStart{Query: gqlSubscribeDevices}
	_, err := g.conn.Subscribe(q, func(m *graphql.Message) {
		p := new(graphql.MessagePayloadData)
		if err := json.Unmarshal(m.Payload, p); err != nil {
			log.Println("GraphQLService: Unable to unmarshal payload:", err)
			return
		}

		r := new(response)
		if err := json.Unmarshal(p.Data, r); err != nil {
			log.Println("GraphQLService: Unable to unmarshal data:", err)
			return
		}
		g.subscribeHandler(r.Devices)
	})

	return err
}

func (g *GraphQLService) SubscribeDevices(f func(devices []*Device)) error {
	g.subscribeHandler = f
	return g.subscribeDevices()
}
