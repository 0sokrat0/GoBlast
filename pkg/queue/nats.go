package queue

import (
	"github.com/nats-io/nats.go"
	"log"
)

type NATSClient struct {
	Conn *nats.Conn
}

func NewNatsClient(url string) (*NATSClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	nc.SetErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		log.Printf("NATS error: %v\n", err)
		log.Printf("Error in NATS connection %s: %s", sub.Subject, err)
	})

	return &NATSClient{
		Conn: nc,
	}, nil
}
