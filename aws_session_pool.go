package main

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
	SESSION_INTERVAL time.Duration = 15 * time.Minute
	SESSION_RETRY    time.Duration = 15 * time.Second
)

var (
	sessMutex sync.RWMutex
	Sessions  = make(map[string]*Session)
)

type Session struct {
	Profile        string
	Region         string
	ChSession      chan bool
	svc            *sqs.SQS
	lastConnection time.Time
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
		ChSession:      make(chan bool, 0),
		lastConnection: time.Now(),
	}

	Sessions[key] = s

	go s.loopSession()

	s.connect()

	return s
}

func (s *Session) loopSession() {

	ti := time.Tick(SESSION_INTERVAL)

	for {
		select {
		case <-ti:
			log.Println("Renew session with AWS", s.lastConnection)
			s.connect()
		case <-s.ChSession:
			if s.lastConnection.Unix() < time.Now().Add(-SESSION_RETRY).Unix() {
				log.Println("Force reconnect")
				s.connect()
			}
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

	s.svc = sqs.New(sess, &aws.Config{Region: aws.String(s.Region)})

	s.lastConnection = time.Now()
}
