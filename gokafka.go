package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/subosito/gotenv"
)

type aivenKafka struct {
	topic     string
	partion   int64
	conn      *kafka.Conn
	host      string
	partition int
}
type kafkaMessage struct {
	UUID        string  `json: "uuid"`
	Timestamp   int64   `json: timestamp"`
	Temperature float64 `json: temperature"`
}

func init() {
	if err := gotenv.Load(); err != nil {
		log.Println("WARN : There is no .env file found. This utility checks for connection details in the System Env.")
	}

}
func main() {

	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

//TODO : function should return slice of errors to be sure the error handling from the defered close doesn't break things.
func run() error {
	p, err := strconv.Atoi(os.Getenv("PARTITION"))
	if err != nil {
		return err
	}
	ak := &aivenKafka{
		topic:     "aiven-topic",
		partion:   0,
		host:      os.Getenv("HOST"),
		partition: p,
	}

	// connect to the kafka service. Read this from our system environment
	ak.conn, err = kafka.DialLeader(context.Background(), "tcp", "localhost:9092", ak.topic, int(ak.partion))
	if err != nil {
		return err
	}
	defer func() {
		err = ak.conn.Close()
	}()

	return nil
}

/*
	sends a temperature reading to the kafka topic
	temp will be random between 0 - 100
*/
func (ak *aivenKafka) writeMsg() error {

	msg := &kafkaMessage{
		UUID:        uuid.New(),
		Timestamp:   time.Now().UnixMilli(),
		Temperature: rand.Float64() * 100}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ak.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = ak.conn.WriteMessages(
		kafka.Message{Value: bytes},
	)
	if err != nil {
		return err
	}
	log.Printf("Sent message\n%v\n", string(bytes))
	return nil
}
