package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"

	"github.com/BurntSushi/toml"
	"github.com/wunderlist/ttlcache"
)

const (
	version = "0.9.2"
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

	reload()

	signal.Notify(chSign, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	go sing()

	<-chMain
	log.Printf("END")
}

func reload() {

	if !debug {
		logwriter, e := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "CARBOnic")
		if e == nil {
			log.SetFlags(0)
			log.SetOutput(logwriter)
		}
	}

	log.Printf("Starting v%s", version)

	mu.Lock()
	closeGroups()
	mu.Unlock()

	var c *Config
	if _, err := toml.DecodeFile(configFile, &c); err != nil {
		log.Printf("ERROR reading config file %s: %s", configFile, err)
		return
	}

	log.Printf("Config file loaded %s", configFile)

	mu.Lock()
	conf = *c
	mu.Unlock()

	for _, g := range conf.Group {
		g.start()
	}

}

func closeGroups() {
	wg := sync.WaitGroup{}
	for _, g := range conf.Group {
		wg.Add(1)
		go func(g *Group) {
			defer wg.Done()
			g.Exit()
		}(g)
	}
	wg.Wait()
}

func sing() {
	for {
		switch <-chSign {
		case syscall.SIGHUP:
			log.Printf("Reloading..")
			reload()
		default:
			log.Printf("Closing...")
			closeGroups()
			chMain <- true
		}
	}
}
