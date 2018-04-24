package sqs

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gabrielperezs/CARBOnic/lib"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	maxNumberOfMessages = 10 // Limit of messages can be read from SQS
	waitTimeSeconds     = 15 // Seconds to keep open the connection to SQS
)

var connPool sync.Map

type config struct {
	URL     string
	Region  string
	Score   int
	Profile string
}

func NewOrGet(c map[string]interface{}) (*SQS, error) {

	cfg := &config{}

	for k, v := range c {
		switch strings.ToLower(k) {
		case "url":
			cfg.URL = v.(string)
		case "region":
			cfg.Region = v.(string)
		case "score":
			cfg.Score = int(v.(int64))
		case "profile":
			cfg.Profile = v.(string)
		}
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("SQS ERROR: URL not found or invalid")
	}

	if cfg.Region == "" {
		return nil, fmt.Errorf("SQS ERROR: Region not found or invalid")
	}

	if cfg.Score <= 0 {
		return nil, fmt.Errorf("SQS ERROR: Score not found or invalid")
	}

	var (
		ok      bool
		havConn interface{}
	)

	if havConn, ok = connPool.Load(cfg.URL); !ok {
		var err error
		havConn, err = newConnection(cfg)
		if err != nil {
			return nil, err
		}
	}

	connPool.Store(cfg.URL, havConn)

	s := havConn.(*SQS)

	return s, nil
}

func newConnection(cfg *config) (interface{}, error) {
	s := &SQS{
		cfg:  cfg,
		done: make(chan bool),
		sess: lib.NewSession(cfg.Profile, cfg.Region),
	}

	go s.listen()

	return s, nil
}

type SQS struct {
	sync.Mutex
	cfg *config

	exiting  bool
	done     chan bool
	groups   sync.Map
	lastMsgs sync.Map
	sess     *lib.Session
}

func (s *SQS) SetGroup(g lib.Group) {
	s.groups.Store(g.GetName(), g)
}

func (s *SQS) DelGroup(g lib.Group) {
	s.groups.Delete(g.GetName())

	// If this resource is not part of a group then finish the thread
	length := 0
	s.groups.Range(func(_, _ interface{}) bool {
		length++
		return false
	})

	if length == 0 {
		s.Exit()
	}
}

func (s *SQS) broadcast(m *lib.Message) {
	s.groups.Range(func(k, v interface{}) bool {
		g := v.(lib.Group)
		g.Chan() <- m
		return true
	})
}

func (s *SQS) StartSession() {
	s.sess = lib.NewSession(s.cfg.Profile, s.cfg.Region)
}

func (s *SQS) HasAlarms() bool {
	length := 0
	s.lastMsgs.Range(func(_, _ interface{}) bool {
		length++
		return false
	})

	if length == 0 {
		return false
	}

	return true
}

func (s *SQS) GetScore() int {
	return s.cfg.Score
}

func (s *SQS) GetLabel() string {
	return s.cfg.URL
}

func (s *SQS) Clean() {
	s.lastMsgs.Range(func(k, i interface{}) bool {
		d := i.(*sqs.DeleteMessageBatchRequestEntry)

		params := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(s.cfg.URL),
			ReceiptHandle: d.ReceiptHandle,
		}

		_, err := s.sess.Svc.DeleteMessage(params)
		if err != nil {
			log.Printf("ERROR cleaning SQS %s [%s]: %v", s.cfg.URL, err, *d.Id)
			return false
		}

		s.lastMsgs.Delete(k.(string))
		log.Printf("Deleted SQS %s %v", s.cfg.URL, *d.Id)
		return true
	})
}

func (s *SQS) Purge() {

	s.Lock()
	s.lastMsgs = sync.Map{}
	s.Unlock()

	params := &sqs.PurgeQueueInput{
		QueueUrl: aws.String(s.cfg.URL),
	}

	// Sending a request using the PurgeQueueRequest method.
	req, resp := s.sess.Svc.PurgeQueueRequest(params)

	err := req.Send()
	if err != nil { // resp is now filled
		log.Println("ERROR sqs purge:", err, resp.String())
		return
	}

	log.Printf("Purged SQS %s", s.cfg.URL)
}

func (s *SQS) storeMessage(msg *sqs.Message) {
	s.lastMsgs.Store(*msg.MessageId, &sqs.DeleteMessageBatchRequestEntry{
		Id:            msg.MessageId,
		ReceiptHandle: msg.ReceiptHandle,
	})
}

func (s *SQS) listen() {
	defer func() {
		log.Printf("SQS: listener ended %s", s.cfg.URL)
	}()

	for {
		s.Lock()
		e := s.exiting
		s.Unlock()

		if e {
			s.done <- true
			close(s.done)
			return
		}

		s.pull()
	}
}

func (s *SQS) pull() error {

	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.cfg.URL),
		MaxNumberOfMessages: aws.Int64(maxNumberOfMessages),
		WaitTimeSeconds:     aws.Int64(waitTimeSeconds),
	}

	resp, err := s.sess.Svc.ReceiveMessage(params)

	if err != nil {
		log.Printf("ERROR: AWS session on %s - %s", s.cfg.URL, err)

		s.broadcast(&lib.Message{
			Score: 5,
			Msg:   fmt.Sprintf("Error reading from sqs: %s", err),
		})

		time.Sleep(15 * time.Second)
		return err
	}

	// Pretty-print the response data.
	for _, msg := range resp.Messages {

		// Add in last messages to be deleted in the next catch
		s.storeMessage(msg)

		body := *msg.Body
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			s.broadcast(&lib.Message{
				Score: s.cfg.Score,
				Msg:   body,
			})
			continue
		}

		if jsonParsed.Exists("Message") {
			intMsg := strings.Replace(jsonParsed.S("Message").String(), "\\\"", "\"", -1)
			intParsed, err := gabs.ParseJSON([]byte(intMsg[1 : len(intMsg)-1]))

			if err != nil {
				s.broadcast(&lib.Message{
					Score: s.cfg.Score,
					Msg:   intMsg,
				})
				continue
			}

			if !intParsed.Exists("AlarmName") {
				log.Printf("Unknown alert format: %s", body)
				s.broadcast(&lib.Message{
					Score: s.cfg.Score,
					Msg:   fmt.Sprintf("Issue JSON: %s", intParsed.StringIndent("", "  ")),
				})
				continue
			}

			s.broadcast(&lib.Message{
				Score: s.cfg.Score,
				Msg: fmt.Sprintf(
					"%s\n%s\n%s\n",
					intParsed.S("AlarmName").String(),
					intParsed.S("AlarmDescription").String(),
					intParsed.S("NewStateReason").String(),
				),
			})
		}

	}

	return nil
}

func (s *SQS) Exit() {
	log.Printf("Inputs: SQS Exiting %s, wait until finish the running requests... (max %ds)", s.GetLabel(), waitTimeSeconds)

	s.Lock()
	s.exiting = true
	s.Unlock()

	// Remove from the main pool
	connPool.Delete(s.cfg.URL)

	<-s.done
	log.Printf("Inputs: SQS - Exit: %s", s.GetLabel())
}
