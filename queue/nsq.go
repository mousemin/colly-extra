package queue

import (
	"github.com/nsqio/go-nsq"
)

type (
	NSQStorage struct {
		Topic    string
		Channel  string
		producer *nsq.Producer
		consumer *nsq.Consumer
		bodyChan chan []byte
	}
	consumerChan struct {
		bytesChan chan []byte
	}
)

func (c *consumerChan) HandleMessage(msg *nsq.Message) error {
	c.bytesChan <- msg.Body
	return nil
}

func NewNSQStorage(producer *nsq.Producer, consumer *nsq.Consumer, topic string) (*NSQStorage, error) {
	storage := &NSQStorage{
		Topic:    topic,
		producer: producer,
		consumer: consumer,
		bodyChan: make(chan []byte),
	}
	storage.consumer.AddHandler(&consumerChan{storage.bodyChan})
	return storage, nil
}

// Init initializes the storage
func (nsq *NSQStorage) Init() error {
	return nil
}

// AddRequest adds a serialized request to the queue
func (nsq *NSQStorage) AddRequest(r []byte) error {
	return nsq.producer.Publish(nsq.Topic, r)
}

// GetRequest pops the next request from the queue
// or returns error if the queue is empty
func (nsq *NSQStorage) GetRequest() ([]byte, error) {
	body := <-nsq.bodyChan
	return body, nil
}

// QueueSize returns with the size of the queue
func (nsq *NSQStorage) QueueSize() (int, error) {
	return 1, nil
}
