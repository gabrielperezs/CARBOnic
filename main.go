package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

const (
	version = "0.2"
)

type Config struct {
	Group []*Group `toml:"Group"`
}

var conf Config

var chMain = make(chan bool)

func main() {

	var config string
	flag.StringVar(&config, "config", "config.toml", "a string var")
	flag.Parse()

	log.Printf("Starting v%s - %s", version, config)

	_, err := toml.DecodeFile(config, &conf)
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}

	for _, g := range conf.Group {
		g.start()
	}

	<-chMain
}
