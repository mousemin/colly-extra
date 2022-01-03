package configurable

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/mousemin/colly-extra/queue"
)

var (
	ErrNotStorage = errors.New("not found configurable storage")
)

type (
	Collector struct {
		mu           *sync.Mutex      // 保证Init唯一
		name         string           // 爬虫名称
		only         bool             // 是否只运行配置爬虫
		storage      Storage          // 配置资源
		queue        queue.Interface  // colly.Queue
		collector    *colly.Collector // colly.Collector
		logger       Logger           // 日志组件
		pipelineFunc PipelineFunc     // 结果pipeline
	}

	PipelineFunc func(name string, v interface{}) error
)

func New(name string, opts ...Option) (*Collector, error) {
	c := &Collector{
		mu:   &sync.Mutex{},
		name: name,
		only: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.storage == nil {
		return nil, ErrNotStorage
	}

	return c, nil
}

func (c *Collector) Name() string {
	return c.name
}

func (c *Collector) PackageName() string {
	return "colly.extra.configurable"
}

func (c *Collector) Init() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.queue == nil {
		c.queue, err = queue.New(DefaultThread, nil)
		if err != nil {
			return
		}
	}
	if c.collector == nil {
		c.collector = colly.NewCollector(colly.DetectCharset())
	}
	if c.logger == nil {
		c.logger = createLogger()
	}

	if c.pipelineFunc == nil {
		c.pipelineFunc = func(name string, v interface{}) error {
			c.logger.Debugf("[result] name: %s send: %v", name, v)
			return nil
		}
	}
	// 注册爬虫的处理函数
	c.collector.OnRequest(c.onRequest)
	c.collector.OnResponse(c.onResponse)
	return
}

func (c *Collector) onRequest(request *colly.Request) {
	// 若只是配置爬虫，没有爬虫配置的标识，直接abort
	confName := request.Ctx.Get(CollyConfName)
	if len(confName) == 0 {
		if c.only {
			request.Abort()
		}
		return
	}
	c.logger.Debugf("request: %s", request.URL.String())
	stepName := request.Ctx.Get(CollyConfStepName)
	if len(stepName) == 0 {
		request.Ctx.Put(CollyConfStepName, CollyConfStepStart)
		conf, err := c.storage.GetConfig(confName)
		if err != nil {
			c.logger.Errorf("request获取配置: %s, err: %s", confName, err.Error())
			return
		}
		request.ProxyURL = conf.GetProxy()
	}
}

func (c *Collector) onResponse(response *colly.Response) {
	urlStr := response.Request.URL.String()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(response.Body))
	if err != nil {
		c.logger.Errorf("解析response.body: %s err: %s", urlStr, err)
		return
	}
	// 配置常量截取
	confName := response.Ctx.Get(CollyConfName)
	conf, err := c.storage.GetConfig(confName)
	if err != nil {
		c.logger.Errorf("response获取配置: %s, err: %s", confName, err.Error())
		return
	}
	stepName := response.Ctx.Get(CollyConfStepName)
	extInfo, _ := response.Ctx.GetAny(CollyConfExt).(map[string]interface{})
	if stepName == CollyConfStepEnd {
		results := conf.GetFinal()
		if len(results) == 0 {
			c.logger.Warnf("获取结果解析配置无效, %s", urlStr)
		}
		pipeline := analyzeDocumentByElements(doc, results, extInfo)
		if err := c.pipelineFunc(confName, pipeline); err != nil {
			c.logger.Warnf("结果pipeline, err: %s", err.Error())
		}
		return
	}
	steps := conf.GetStep(stepName)
	if len(steps) == 0 {
		c.logger.Warnf("获取步骤: %s 解析配置无效, %s", stepName, urlStr)
		return
	}

	for i, step := range steps {
		if len(step.CSSPath) == 0 {
			c.logger.Warnf("步骤: %s-%d selector 不存在", stepName, i)
			continue
		}
		ctx := colly.NewContext()
		ctx.Put(CollyConfStepName, step.Next)
		ctx.Put(CollyConfName, confName)
		if len(step.Ext) != 0 {
			ext := analyzeDocumentByElements(doc, step.Ext, extInfo)
			ctx.Put(CollyConfExt, ext)
		}
		doc.Find(step.CSSPath).Each(func(i int, e *goquery.Selection) {
			if val, ok := e.Attr(step.GetAttr()); ok {
				hrefURL := response.Request.AbsoluteURL(val)
				if len(hrefURL) == 0 {
					return
				}
				u, err := url.Parse(hrefURL)
				if err != nil {
					c.logger.Debugf("%s 解析失败, err: %s", urlStr, err)
					return
				}
				err = c.queue.AddRequest(&colly.Request{
					URL:      u,
					Ctx:      ctx,
					Method:   step.GetHttpMethod(),
					ProxyURL: conf.GetProxy(),
				})
				if err != nil {
					c.logger.Errorf("添加request %s err: %s", urlStr, err)
				}
			}
		})

	}
}

func (c *Collector) Start() error {
	return c.queue.Run(c.collector)
}

func (c *Collector) Stop() error {
	return nil
}

func (c *Collector) GracefulStop(ctx context.Context) error {
	return nil
}
