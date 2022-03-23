package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

type kafkaMessage struct {
	Timestamp   string  `json:"timestamp"`
	Temperature float64 `json:"temperature"`
}
type credentials struct {
	ServiceCertPath string
	ServiceKeyPath  string
	CaCertPath      string
	CaCert          []byte
	KeyPair         tls.Certificate
}

func init() {
	if e := godotenv.Load(); e != nil {
		log.Printf("%v", e)
	}
}

func main() {
	c, err := loadCredentials()
	if err != nil {
		log.Fatal(err)
	}
	if err := run(c); err != nil {
		log.Fatalf("%v", err)
	}
}

func loadCredentials() (*credentials, error) {

	c := &credentials{}
	c.ServiceCertPath = os.Getenv("CERT_PATH")
	c.ServiceKeyPath = os.Getenv("KEY_PATH")
	c.CaCertPath = os.Getenv("CA_CERT_PATH")

	var err error
	c.KeyPair, err = tls.LoadX509KeyPair(c.ServiceCertPath, c.ServiceKeyPath)
	if err != nil {
		return nil, err
	}

	c.CaCert, err = ioutil.ReadFile(c.CaCertPath)
	if err != nil {
		log.Printf("unable to find path %s\n", c.CaCertPath)
		return nil, err
	}
	return c, nil
}

//TODO : function should return slice of errors to be sure the error handling from the defered close doesn't break things.
func run(c *credentials) error {

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(c.CaCert)
	if !ok {
		return fmt.Errorf("Failed to parse the CA Certificate file at : %s", c.CaCertPath)
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS: &tls.Config{
			Certificates: []tls.Certificate{c.KeyPair},
			RootCAs:      caCertPool,
		},
	}

	producer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("HOST_URL")},
		Topic:   os.Getenv("TOPIC"),
		Dialer:  dialer,
	})
	// TODO : Handle the error this produces
	defer producer.Close()

	// yes, this sensor will run forever!!
	var err error
	for {
		if err = writeMsg(producer); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}

/*
	sends a temperature reading to the kafka topic
	temp will be random between 0 - 100
*/
func writeMsg(p *kafka.Writer) error {

	t := time.Now()

	msg := &kafkaMessage{
		Timestamp:   t.Format("2006-01-02T15:04:05-0700"),
		Temperature: rand.Float64() * 100}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = p.WriteMessages(context.Background(), kafka.Message{Key: []byte(uuid.New().String()), Value: bytes})

	if err != nil {
		return err
	}

	log.Printf("Sent message\n%v\n", string(bytes))
	return nil
}
