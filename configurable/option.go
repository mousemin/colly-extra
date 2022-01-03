package configurable

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

type Option func(c *Collector)

func WithOnly(only bool) Option {
	return func(c *Collector) {
		c.only = only
	}
}

func WithConfigStorage(storage Storage) Option {
	return func(c *Collector) {
		c.storage = storage
	}
}

func WithCollector(collector *colly.Collector) Option {
	return func(c *Collector) {
		c.collector = collector
	}
}

func WithQueue(q *queue.Queue) Option {
	return func(c *Collector) {
		c.queue = q
	}
}

func WithLogger(l Logger) Option {
	return func(c *Collector) {
		c.logger = l
	}
}

func WithPipeline(pipeLineFunc PipelineFunc) Option {
	return func(c *Collector) {
		c.pipelineFunc = pipeLineFunc
	}
}
