package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"

	"github.com/BurntSushi/toml"
	"github.com/spaolacci/murmur3"
	"github.com/wunderlist/ttlcache"
)

const (
	version = "0.4.0"
)

type Config struct {
	Group []*Group `toml:"Group"`
}

var (
	conf       Config
	configFile string

	debug = false

	chSign = make(chan os.Signal, 10)
	chMain = make(chan bool)
	cache  = ttlcache.NewCache(time.Millisecond * 2500)

	mu sync.Mutex
)

func main() {

	flag.StringVar(&configFile, "config", "config.toml", "Configuration file")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	log.Printf("Starting v%s", version)

	reload()

	signal.Notify(chSign, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	go sing()

	<-chMain
	log.Printf("END")
}

func reload() {

	mu.Lock()
	for _, g := range conf.Group {
		g.Exit()
	}
	mu.Unlock()

	var c *Config
	if _, err := toml.DecodeFile(configFile, &c); err != nil {
		log.Printf("ERROR reading config file %s: %s", configFile, err)
		return
	}

	log.Printf("Config file loaded %s", configFile)

	mu.Lock()
	conf = *c
	if !debug {
		log.SetFlags(0)
		log.SetOutput(new(logWriter))
	}
	mu.Unlock()

	for _, g := range conf.Group {
		g.start()
	}

}

func sing() {
	for {
		switch <-chSign {
		case syscall.SIGHUP:
			log.Printf("Reloading..")
			reload()
		default:
			for _, g := range conf.Group {
				g.Exit()
			}
			log.Printf("Closing by signal")
			chMain <- true
		}
	}
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
