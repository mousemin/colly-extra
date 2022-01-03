package queue

import (
	"net/url"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

type Interface interface {
	IsEmpty() bool
	AddURL(URL string) error
	AddRequest(r *colly.Request) error
	Size() (int, error)
	Run(c *colly.Collector) error
}

// Queue is a request queue which uses a Collector to consume
// requests in multiple threads
type Queue struct {
	// Threads defines the number of consumer threads
	Threads int
	storage queue.Storage
	wake    chan struct{}
	mut     sync.Mutex // guards wake
}

// New creates a new queue with a Storage specified in argument
// A standard InMemoryQueueStorage is used if Storage argument is nil.
func New(threads int, s queue.Storage) (*Queue, error) {
	if s == nil {
		s = &queue.InMemoryQueueStorage{MaxSize: 100000}
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	return &Queue{
		Threads: threads,
		storage: s,
	}, nil
}

// IsEmpty returns true if the queue is empty
func (q *Queue) IsEmpty() bool {
	s, _ := q.Size()
	return s == 0
}

// AddURL adds a new URL to the queue
func (q *Queue) AddURL(URL string) error {
	u, err := url.Parse(URL)
	if err != nil {
		return err
	}
	r := &colly.Request{
		URL:    u,
		Method: "GET",
	}
	return q.AddRequest(r)
}

// AddRequest adds a new Request to the queue
func (q *Queue) AddRequest(r *colly.Request) error {
	d, err := r.Marshal()
	if err != nil {
		return err
	}
	return q.storage.AddRequest(d)
}

// Size returns the size of the queue
func (q *Queue) Size() (int, error) {
	return q.storage.QueueSize()
}

// Run starts consumer threads and calls the Collector
// to perform requests. Run blocks while the queue has active requests
// The given Storage must not be used directly while Run blocks.
func (q *Queue) Run(c *colly.Collector) error {
	q.mut.Lock()
	if q.wake != nil {
		q.mut.Unlock()
		panic("cannot call duplicate Queue.Run")
	}
	q.wake = make(chan struct{})
	q.mut.Unlock()

	requestc := make(chan *colly.Request)
	complete, errc := make(chan struct{}), make(chan error, 1)
	for i := 0; i < q.Threads; i++ {
		go independentRunner(requestc, complete)
	}
	go q.loop(c, requestc, complete, errc)
	defer close(requestc)
	return <-errc
}

func (q *Queue) loop(c *colly.Collector, requestc chan<- *colly.Request, complete <-chan struct{}, errc chan<- error) {
	var active int
	for {
		size, err := q.storage.QueueSize()
		if err != nil {
			errc <- err
			break
		}
		if size == 0 && active == 0 {
			// Terminate when
			//   1. No active requests
			//   2. Emtpy queue
			errc <- nil
			break
		}
		sent := requestc
		var req *colly.Request
		if size > 0 {
			req, err = q.loadRequest(c)
			if err != nil {
				// ignore an error returned by GetRequest() or
				// UnmarshalRequest()
				continue
			}
		} else {
			sent = nil
		}
	Sent:
		for {
			select {
			case sent <- req:
				active++
				break Sent
			case <-q.wake:
				if sent == nil {
					break Sent
				}
			case <-complete:
				active--
				if sent == nil && active == 0 {
					break Sent
				}
			}
		}
	}
}

func independentRunner(requestc <-chan *colly.Request, complete chan<- struct{}) {
	for req := range requestc {
		req.Do()
		complete <- struct{}{}
	}
}

func (q *Queue) loadRequest(c *colly.Collector) (*colly.Request, error) {
	buf, err := q.storage.GetRequest()
	if err != nil {
		return nil, err
	}
	copied := make([]byte, len(buf))
	copy(copied, buf)
	return c.UnmarshalRequest(copied)
}
