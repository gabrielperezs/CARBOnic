package main

import (
	"flag"
	"fmt"
	"log"

	"time"

	"github.com/BurntSushi/toml"
	"github.com/spaolacci/murmur3"
	"github.com/wunderlist/ttlcache"
)

const (
	version = "0.3.1"
)

type Config struct {
	Group []*Group `toml:"Group"`
}

var (
	conf Config

	debug = false

	chMain = make(chan bool)
	cache  = ttlcache.NewCache(time.Millisecond * 2500)
)

func main() {

	var config string
	flag.StringVar(&config, "config", "config.toml", "a string var")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	if !debug {
		log.SetFlags(0)
		log.SetOutput(new(logWriter))
	}

	log.Printf("Starting v%s - %s", version, config)

	_, err := toml.DecodeFile(config, &conf)
	if err != nil {
		// handle error
		log.Panic(err)
		return
	}

	for _, g := range conf.Group {
		g.start()
	}

	<-chMain
}

func ignoreDup(key string) bool {
	hash := fmt.Sprint(murmur3.Sum32([]byte(key)))

	var found bool

	_, found = cache.Get(hash)
	if found {
		return false
	}

	cache.Set(hash, "1")

	return true
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}
