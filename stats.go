package main

import (
	"github.com/pagerduty/godspeed"
	"log"
	"os"
	"strconv"
	"strings"
)

var StatsClient *godspeed.AsyncGodspeed

var StatsEnabled = false

func InitStats() error {
	log.Println("Initializing statsd: " + os.Getenv("DATADOG_PORT_8125_UDP_ADDR") + ":" + os.Getenv("DATADOG_PORT_8125_UDP_PORT"))
	port := os.Getenv("DATADOG_PORT_8125_UDP_PORT")
	intPort, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return err
	}
	client, err := godspeed.NewAsync(os.Getenv("DATADOG_PORT_8125_UDP_ADDR"), int(intPort), true)
	if err != nil {
		return err
	}
	StatsClient = client
	StatsEnabled = true

	StatsClient.SetNamespace(os.Getenv("DATADOG_STATS_NAMESPACE"))

	tags := os.Getenv("DATADOG_STATS_TAGS")
	if len(tags) > 0 {
		tag := strings.Split(tags, ",")
		StatsClient.AddTags(tag)
	}
	return nil

}
