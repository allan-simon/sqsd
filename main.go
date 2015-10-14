package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/codegangsta/cli"
	"log"
	"os"
	"strings"
	"time"
)

var client *sqs.SQS

type config struct {
	queueUrl        string
	deadQueueUrl    string
	workerUrl       string
	timeout         int
	parallel        int
	elastic         bool
	maxReceiveCount int
}

var workerConfig config

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = "sqsd"
	app.Usage = "POST messages from SQS (or ElasticMQ) to an endpoint"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "accessKey",
			Usage:  "AWS access key (needs SQS read/write permissions)",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secretKey",
			Usage:  "AWS secret key",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "AWS region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
		cli.StringFlag{
			Name:   "queue",
			Usage:  "URL of the message queue",
			EnvVar: "QUEUE_URL",
		},
		cli.BoolFlag{
			Name:   "elastic",
			Usage:  "Set to true to use ElasticMQ linked via Docker as \"queue\" rather than SQS",
			EnvVar: "ELASTIC",
		},
		cli.StringFlag{
			Name:   "host",
			Usage:  "Host to POST messages to",
			EnvVar: "WORKER_PORT_8080_TCP_ADDR",
		},
		cli.StringFlag{
			Name:   "port",
			Usage:  "Port on the host",
			EnvVar: "WORKER_PORT_8080_TCP_PORT",
		},
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "Endpoint/path on the host",
			Value:  "/",
			EnvVar: "ENDPOINT",
		},
		cli.IntFlag{
			Name:   "timeout",
			Usage:  "Maximum time to wait (in seconds) before marking a message as 'failed'",
			Value:  30,
			EnvVar: "TIMEOUT",
		},
		cli.IntFlag{
			Name:   "parallel",
			Usage:  "Maximum number of concurrent requests to make to the host",
			Value:  30,
			EnvVar: "PARALLEL",
		},
		cli.IntFlag{
			Name:   "maxReceiveCount",
			Usage:  "If using ElasticMQ, this is the maximum number of times a message will be tried before being sent to the dead letter queue. In SQS this happens automatically if you set up the dead letter queue",
			EnvVar: "MAX_RECEIVE_COUNT",
		},
		cli.StringFlag{
			Name:   "deadQueue",
			Usage:  "If using ElasticMQ, URL of the dead letter queue you want to use. Is not used if using SQS",
			EnvVar: "DEAD_QUEUE_URL",
		},
	}

	app.Action = func(c *cli.Context) {

		workerConfig = config{
			queueUrl:        c.String("queue"),
			deadQueueUrl:    c.String("deadQueue"),
			maxReceiveCount: c.Int("maxReceiveCount"),
			workerUrl:       "http://" + c.String("host") + ":" + c.String("port") + c.String("endpoint"),
			timeout:         c.Int("timeout"),
			parallel:        c.Int("parallel"),
		}

		// For simplicity, set the environment variables to the flags even if they came originally from the environment vars.
		// That way we can just use AWS' default credentials provider
		os.Setenv("AWS_ACCESS_KEY_ID", c.String("accessKey"))
		os.Setenv("AWS_SECRET_ACCESS_KEY", c.String("secretKey"))

		if c.Bool("elastic") {
			workerConfig.elastic = true
			client = sqs.New(&aws.Config{
				Endpoint: aws.String("http://" + os.Getenv("QUEUE_PORT_9324_TCP_ADDR") + ":" + os.Getenv("QUEUE_PORT_9324_TCP_PORT")),
				Region:   aws.String(c.String("region")),
			})
		} else {
			client = sqs.New(&aws.Config{
				Region: aws.String(c.String("region")),
			})
		}

		log.Println("Making sure queue exists...")
		spl := strings.Split(workerConfig.queueUrl, "/")
		_, err := client.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(spl[len(spl)-1]),
		})
		if err != nil {
			log.Fatal(err)
		}

		if workerConfig.elastic {
			spl = strings.Split(workerConfig.deadQueueUrl, "/")
			_, err = client.CreateQueue(&sqs.CreateQueueInput{
				QueueName: aws.String(spl[len(spl)-1]),
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		// Just to test...
		log.Println(client.SendMessage(&sqs.SendMessageInput{
			MessageBody: aws.String(`{"foo":"bar"}`),
			QueueUrl:    aws.String(workerConfig.queueUrl),
		}))

		work()

	}

	app.Run(os.Args)
}

func work() {
	workerChannel := make(chan Worker, workerConfig.parallel)
	for i := 0; i < workerConfig.parallel; i++ {
		workerChannel <- Worker{}
	}

	for {
		select {
		case w := <-workerChannel:
			go w.Work(workerChannel)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
