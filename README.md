sqsd
============

This is a port of the `sqsd` bundled in AWS Elastic Beanstalk's worker tier, meant to be run as a **Docker Container** linked to the worker container and optionally a queue container (if you want to use [ElasticMQ](https://github.com/adamw/elasticmq))

It reads from an SQS-compatible queue and forwards messages to a specified URL in a `POST` request. It deletes the message from the queue if the endpoint returns a 200-level response, otherwise it leaves it in the queue to be processed again. To prevent infinite processing of bad messages, you should create a dead-letter queue if you're using SQS. If using ElasticMQ, you can make use of the built-in dead letter queue function to move repeatedly failing messages into the designated dead letter queue.

## Usage
All parameters can be set either with flags or with environment variables (in `[$BRACKETS]`). Environment variables for `queue` and `host` parameters

```
NAME:
   sqsd - POST messages from SQS (or ElasticMQ) to an endpoint

USAGE:
   sqsd [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --accessKey 			AWS access key (needs SQS read/write permissions) [$AWS_ACCESS_KEY_ID]
   --secretKey 			AWS secret key [$AWS_SECRET_ACCESS_KEY]
   --region 			AWS region [$AWS_DEFAULT_REGION]
   --queue 			URL of the message queue [$QUEUE_URL]
   --elastic			Set to true to use ElasticMQ linked via Docker as "queue" rather than SQS [$ELASTIC]
   --host 			Host to POST messages to [$WORKER_PORT_8080_TCP_ADDR]
   --port 			Port on the host [$WORKER_PORT_8080_TCP_PORT]
   --endpoint "/"		Endpoint/path on the host [$ENDPOINT]
   --timeout "30"		Maximum time to wait (in seconds) before marking a message as 'failed' [$TIMEOUT]
   --parallel "30"		Maximum number of concurrent requests to make to the host [$PARALLEL]
   --maxReceiveCount "0"	If using ElasticMQ, this is the maximum number of times a message will be tried before being sent to the dead letter queue. In SQS this happens automatically if you set up the dead letter queue [$MAX_RECEIVE_COUNT]
   --deadQueue 			If using ElasticMQ, URL of the dead letter queue you want to use. Is not used if using SQS [$DEAD_QUEUE_URL]
   --help, -h			show help
   --version, -v		print the version
```

### With Docker
See the [sample docker-compose](docker-compose.sample.yml) file.

Image is automatically built on push to `dispatch/sqsd:latest`
