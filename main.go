package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"
)

const (
	version = "0.1"
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
