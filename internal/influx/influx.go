package influx

import (
	"context"
	"encoding/json"
	"log"

	"github.com/chiaf1/mqttingest_solar/internal/config"
	"github.com/chiaf1/mqttingest_solar/internal/ingestion"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// MQTT payload for energy data
type EnergyPayload struct {
	Production  float64 `json:"production"`
	Consumption float64 `json:"consumption"`
}

// Worker function to handle mqtt sub requests for data ingestion into an influx db
func InputToDB(client influxdb2.Client, conf config.InfluxConfig, msg ingestion.Message) {
	switch msg.Topic {
	case "energy/data":
		writeEnergy(client, conf, msg)
		return
	default:
		log.Printf("[influx_ingest] Unknown topic [%s], payload:\n%s", msg.Topic, string(msg.Payload))
	}
}

// Handler function for energy data, it parses the payload and writes to the influxdb
func writeEnergy(client influxdb2.Client, conf config.InfluxConfig, msg ingestion.Message) {
	// data parsing from mqtt payload
	var data EnergyPayload
	err := json.Unmarshal(msg.Payload, &data)
	if err != nil {
		log.Printf("[influx_ingest] error parsing JSON: %v", err)
		return
	}

	// data point preparation for influx db
	tags := map[string]string{
		"source": "arduino",
	}
	fields := map[string]interface{}{
		"production":  data.Production,
		"consumption": data.Consumption,
	}
	p := influxdb2.NewPoint("energy", tags, fields, msg.Time)

	// writing to influx db
	writeAPI := client.WriteAPIBlocking(conf.Org, conf.Bucket)

	err = writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		log.Printf("[influx_ingest] error writing to influx: %v", err)
	}
}
