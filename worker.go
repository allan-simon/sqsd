package main

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Worker struct{}

func (w Worker) Work(ch chan Worker) {
	defer func() {
		ch <- w
	}()

	log.Println("Requesting message")
	response, err := client.ReceiveMessage(&sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(1),
		QueueUrl:            aws.String(workerConfig.queueUrl),
		// QueueName:           aws.String(workerConfig.queueName),
		WaitTimeSeconds:   aws.Int64(20),
		VisibilityTimeout: aws.Int64(int64(workerConfig.timeout + 5)),
		AttributeNames: []*string{
			aws.String("ApproximateReceiveCount"),
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	if len(response.Messages) > 0 {
		msg := response.Messages[0]
		log.Printf("%+v\n\n", msg)
		err = w.handleMessage(msg)
		if err != nil {
			// It'll be picked up again after the visibility timeout. If we're on ElasticMQ, which doesn't have dead letter queue support, we need to delete the message if it reaches the maximum receive count.
			log.Println(err)

			if workerConfig.elastic {
				if receiveCount, ok := msg.Attributes["ApproximateReceiveCount"]; ok {
					parsed, err := strconv.ParseInt(*receiveCount, 10, 64)
					if err != nil {
						log.Println(err)
					} else {
						if parsed >= int64(workerConfig.maxReceiveCount) {
							log.Println("Sending message to dead letter queue")
							_, err = client.SendMessage(&sqs.SendMessageInput{
								MessageBody: msg.Body,
								QueueUrl:    aws.String(workerConfig.deadQueueUrl),
							})
							if err != nil {
								log.Println(err)
							} else {
								err = deleteMessage(msg)
								if err != nil {
									log.Println(err)
								}
							}

						}
					}
				}
			}
		} else {
			// Delete from the queue
			err = deleteMessage(msg)

			if err != nil {
				log.Println(err)
			}
		}
	}
}

func deleteMessage(msg *sqs.Message) error {
	_, err := client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(workerConfig.queueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	})
	return err
}

func (Worker) handleMessage(msg *sqs.Message) error {
	client := http.Client{
		Timeout: time.Duration(time.Duration(workerConfig.timeout) * time.Second),
	}

	body := *msg.Body

	reader := bytes.NewReader([]byte(body))

	response, err := client.Post(workerConfig.workerUrl, "application/json", reader)

	if err != nil {
		return err
	}

	if response.StatusCode > 299 || response.StatusCode < 200 {
		return errors.New("Host returned error status (" + response.Status + ")")
	} else {
		return nil
	}
}
