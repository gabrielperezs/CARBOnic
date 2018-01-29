package lib

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)

const (
	dupExpiration    = 10 * time.Second
	dupCleanInterval = 5 * time.Minute
)

var (
	localCache = &sync.Map{}
)

func init() {
	go func() {
		ticker := time.NewTicker(dupCleanInterval)
		for _ = range ticker.C {
			cleanChech(time.Now())
		}
	}()
}

func cleanChech(n time.Time) {
	localCache.Range(func(k, v interface{}) bool {
		if n.After(v.(time.Time).Add(dupExpiration)) {
			log.Printf("Auto-Deleted %s, %s - %s", k.(string), v.(time.Time).Add(dupExpiration), n)
			localCache.Delete(k.(string))
		}
		return true
	})
}

func IsDupMessage(key string, m *Message) bool {
	mur := murmur3.Sum32([]byte(m.Msg))
	k := fmt.Sprintf("%s_%d_%d", key, mur, m.Score)

	n := time.Now()
	cleanChech(n)

	_, ok := localCache.Load(k)
	if !ok {
		log.Printf("Stored %s", k)
		localCache.Store(k, n)
		return false
	}

	log.Printf("Dup %s", k)
	localCache.Store(k, n)
	return true
}
