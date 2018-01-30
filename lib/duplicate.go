package lib

import (
	"fmt"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)

const (
	dupExpiration = 5 * time.Second
)

var (
	localCache = &sync.Map{}
)

func cleanChech(n time.Time) {
	localCache.Range(func(k, v interface{}) bool {
		if n.After(v.(time.Time).Add(dupExpiration)) {
			localCache.Delete(k.(string))
		}
		return true
	})
}

func IsDupMessage(key, m string) bool {
	mur := murmur3.Sum32([]byte(m))
	k := fmt.Sprintf("%s_%d", key, mur)

	n := time.Now()
	cleanChech(n)

	_, ok := localCache.LoadOrStore(k, n)
	if ok {
		//log.Printf("Dup %s %d:%d %d:%d", k, oldN.(time.Time).Minute(), oldN.(time.Time).Second(), n.Minute(), n.Second())
		return true
	}

	return false
}
