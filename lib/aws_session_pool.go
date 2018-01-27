package lib

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	SESSION_INTERVAL time.Duration = 1 * time.Hour
	SESSION_RETRY    time.Duration = 30 * time.Second
)

var (
	sessMutex sync.RWMutex
	Sessions  = make(map[string]*Session)
)

type Session struct {
	Profile        string
	Region         string
	Svc            *sqs.SQS
	lastConnection time.Time
	tick           *time.Ticker
	done           chan bool
}

func NewSession(profile string, region string) *Session {

	sessMutex.RLock()
	defer sessMutex.RUnlock()

	key := fmt.Sprintf("%s:%s", profile, region)

	if sqs, ok := Sessions[key]; ok {
		return sqs
	}

	s := &Session{
		Profile:        profile,
		Region:         region,
		lastConnection: time.Now(),
		tick:           time.NewTicker(SESSION_INTERVAL),
		done:           make(chan bool, 0),
	}

	Sessions[key] = s

	go s.loopSession()

	s.connect()

	return s
}

func (s *Session) loopSession() {

	defer s.tick.Stop()

	for {
		select {
		case <-s.tick.C:
			log.Println("Renew session with AWS", s.lastConnection)
			s.connect()
		case <-s.done:
			return
		}
	}
}

func (s *Session) connect() {

	opt := session.Options{}

	if s.Profile != "" {
		opt = session.Options{
			Profile: s.Profile,
		}
	}

	sess, err := session.NewSessionWithOptions(opt)

	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	s.Svc = sqs.New(sess, &aws.Config{Region: aws.String(s.Region)})

	s.lastConnection = time.Now()
}
