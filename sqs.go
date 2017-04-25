package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var _ time.Duration
var _ bytes.Buffer

type SQS struct {
	Url     string
	Region  string
	Profile string
	Score   int

	sess *Session
}

func (s *SQS) purge() {

	sess := NewSession(s.Profile, s.Region)

	params := &sqs.PurgeQueueInput{
		QueueUrl: aws.String(s.Url),
	}

	// Example sending a request using the PurgeQueueRequest method.
	req, resp := sess.svc.PurgeQueueRequest(params)

	err := req.Send()
	if err != nil { // resp is now filled
		log.Println("ERROR sqs purge:", err, resp.String())
	}

}

func (s *SQS) pullSQS(ch chan *Message) {

	sess := NewSession(s.Profile, s.Region)

	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.Url),
		MaxNumberOfMessages: aws.Int64(10),
	}
	resp, err := sess.svc.ReceiveMessage(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.

		log.Println("Error", err)

		ch <- &Message{
			score: 0,
			msg:   fmt.Sprintf("Error reading from sqs: %s", err),
		}
		return
	}

	// Pretty-print the response data.
	for _, msg := range resp.Messages {
		body := *msg.Body
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			ch <- &Message{
				score: s.Score,
				msg:   body,
			}
			return
		}

		if jsonParsed.Exists("Message") {
			intMsg := strings.Replace(jsonParsed.S("Message").String(), "\\\"", "\"", -1)
			intParsed, err := gabs.ParseJSON([]byte(intMsg[1 : len(intMsg)-1]))
			if err != nil {
				ch <- &Message{
					score: s.Score,
					msg:   intMsg,
				}
				return
			}

			ch <- &Message{
				score: s.Score,
				msg: fmt.Sprintf(
					"%s: %s %s",
					intParsed.S("AlarmName").String(),
					intParsed.S("AlarmDescription").String(),
					intParsed.S("NewStateReason").String(),
				),
			}
			return
		}

	}
}
