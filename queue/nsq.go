package queue

import (
	"time"

	"github.com/nsqio/go-nsq"
)

type NSQQueue struct {
	config   NSQQueueConfig
	producer *nsq.Producer
	consumer *nsq.Consumer
	bodyChan chan []byte
}

type NSQQueueConfig struct {
	Topic        string
	Channel      string
	ProducerAddr string
	ConsumerAddr string
}

type consumerChan struct {
	bytesChan chan []byte
}

func (c *consumerChan) HandleMessage(msg *nsq.Message) error {
	c.bytesChan <- msg.Body
	return nil
}

func NewNSQQueue(config NSQQueueConfig) (*NSQQueue, error) {
	cfg := nsq.NewConfig()
	nsqClient := &NSQQueue{
		config:   config,
		bodyChan: make(chan []byte),
	}
	var err error
	cfg.LookupdPollInterval = time.Second //设置重连时间
	nsqClient.producer, err = nsq.NewProducer(config.ProducerAddr, cfg)
	if err != nil {
		return nil, err
	}
	c, err := nsq.NewConsumer(config.Topic, config.Channel, cfg)
	if err != nil {
		panic(err)
	}
	c.SetLogger(nil, nsq.LogLevelWarning)
	c.AddHandler(&consumerChan{nsqClient.bodyChan})
	if err := c.ConnectToNSQLookupd(config.ConsumerAddr); err != nil {
		panic(err)
	}
	return nsqClient, nil
}

// Init initializes the storage
func (nsq *NSQQueue) Init() error {
	return nil
}

// AddRequest adds a serialized request to the queue
func (nsq *NSQQueue) AddRequest(r []byte) error {
	return nsq.producer.Publish(nsq.config.Topic, r)
}

// GetRequest pops the next request from the queue
// or returns error if the queue is empty
func (nsq *NSQQueue) GetRequest() ([]byte, error) {
	body := <-nsq.bodyChan
	return body, nil
}

// QueueSize returns with the size of the queue
func (nsq *NSQQueue) QueueSize() (int, error) {
	return 1, nil
}
