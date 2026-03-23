package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chiaf1/mqttingest_solar/internal/config"
	"github.com/chiaf1/mqttingest_solar/internal/influx"
	"github.com/chiaf1/mqttingest_solar/internal/ingestion"
	"github.com/chiaf1/mqttingest_solar/internal/mqtt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const CONFIG_PATH = "./config.yaml"

func main() {
	// Load configs from file
	var conf config.Config
	err := conf.Load(CONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}
	err = conf.Validate()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config Loaded")

	// MQTT connection
	// Client creation
	client := mqtt.NewClient(conf.MQTT.Broker, conf.MQTT.ClientID, conf.MQTT.QoS, conf.MQTT.ConnectionInterval, conf.MQTT.Topics)
	// First connection attempt
	err = mqtt.FirstConnect(client, conf.MQTT.MaxRetry, conf.MQTT.ConnectionInterval, conf.MQTT.MaxDelay)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Client connected")

	// Influx db client creation
	infClient := influxdb2.NewClient(conf.InfluxDB.Url, conf.InfluxDB.Token)
	// handler function for managing messages using the client just created
	handler := func(msg ingestion.Message) {
		influx.InputToDB(infClient, conf.InfluxDB, msg)
	}

	// Start ingestion workers
	ingestion.StartWorkers(5, handler)

	// Grace full shut down
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	log.Println("Service shutdown started...")

	// Closing the channel, will stop all workers
	close(ingestion.WorkerInput)

	// Closing MQTT connection
	client.Unsubscribe(conf.MQTT.Topics...)
	client.Disconnect(250)
	time.Sleep(500 * time.Millisecond)

	// Closing influx db connection
	infClient.Close()

	log.Println("Program ended")
}
