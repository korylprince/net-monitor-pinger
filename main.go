package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	c := new(config)
	envconfig.MustProcess("", c)

	_, err := NewManager(c)
	if err != nil {
		log.Fatalln("Unable to create manager:", err)
	}

	select {}
}
