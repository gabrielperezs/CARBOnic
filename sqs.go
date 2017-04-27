package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"sync"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var _ time.Duration
var _ bytes.Buffer

type SQS struct {
	sync.Mutex
	Url     string
	Region  string
	Profile string
	Score   int

	lastMsgs []*sqs.DeleteMessageBatchRequestEntry
	sess     *Session
}

func (s *SQS) startSession() {
	s.sess = NewSession(s.Profile, s.Region)
}

func (s *SQS) hasAlarms() bool {
	s.Lock()
	defer s.Unlock()

	if len(s.lastMsgs) <= 0 {
		return false
	}

	return true
}

func (s *SQS) clean() {

	s.Lock()
	defer s.Unlock()

	var noDeleted []*sqs.DeleteMessageBatchRequestEntry

	for _, d := range s.lastMsgs {
		params := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(s.Url),
			ReceiptHandle: d.ReceiptHandle,
		}

		_, err := s.sess.svc.DeleteMessage(params)
		ID := *d.Id
		if err != nil {
			log.Printf("ERROR cleaning SQS %s [%s]: %v", s.Url, err, ID)
			noDeleted = append(noDeleted, d)
			continue
		}

		log.Printf("Deleted SQS %s %v", s.Url, ID)

	}

	s.lastMsgs = noDeleted

}

func (s *SQS) purge() {

	s.Lock()
	defer s.Unlock()

	s.lastMsgs = nil

	params := &sqs.PurgeQueueInput{
		QueueUrl: aws.String(s.Url),
	}

	// Example sending a request using the PurgeQueueRequest method.
	req, resp := s.sess.svc.PurgeQueueRequest(params)

	err := req.Send()
	if err != nil { // resp is now filled
		log.Println("ERROR sqs purge:", err, resp.String())
	}

}

func (s *SQS) storeMessage(msg *sqs.Message) {
	s.Lock()
	defer s.Unlock()

	// Add in last messages to be deleted in the next catch
	for _, v := range s.lastMsgs {
		if v.Id == msg.MessageId {
			return
		}
	}

	s.lastMsgs = append(s.lastMsgs, &sqs.DeleteMessageBatchRequestEntry{
		Id:            msg.MessageId,
		ReceiptHandle: msg.ReceiptHandle,
	})
}

func (s *SQS) pullSQS(ch chan *Message) {

	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.Url),
		MaxNumberOfMessages: aws.Int64(5),
	}
	resp, err := s.sess.svc.ReceiveMessage(params)

	if err != nil {
		log.Println("Error", err)

		ch <- &Message{
			score: 5,
			msg:   fmt.Sprintf("Error reading from sqs: %s", err),
		}
		return
	}

	// Pretty-print the response data.
	for _, msg := range resp.Messages {

		// Add in last messages to be deleted in the next catch
		s.storeMessage(msg)

		body := *msg.Body
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			ch <- &Message{
				score: s.Score,
				msg:   body,
			}
			continue
		}

		if jsonParsed.Exists("Message") {
			intMsg := strings.Replace(jsonParsed.S("Message").String(), "\\\"", "\"", -1)
			intParsed, err := gabs.ParseJSON([]byte(intMsg[1 : len(intMsg)-1]))

			if err != nil {
				ch <- &Message{
					score: s.Score,
					msg:   intMsg,
				}
				continue
			}

			if !intParsed.Exists("AlarmName") {
				log.Println("Unknown alert format:", body)
				ch <- &Message{
					score: s.Score,
					msg:   fmt.Sprintf("Issue JSON: %s", intParsed.StringIndent("", "  ")),
				}
				continue
			}

			ch <- &Message{
				score: s.Score,
				msg: fmt.Sprintf(
					"%s\n%s\n%s\n",
					intParsed.S("AlarmName").String(),
					intParsed.S("AlarmDescription").String(),
					intParsed.S("NewStateReason").String(),
				),
			}

		}

	}
}
