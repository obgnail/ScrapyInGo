package engine

import (
	"log"
	"time"

	"github.com/juju/errors"

	"github.com/obgnail/ScrapyInGo/core/downloader"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/middleware"
	"github.com/obgnail/ScrapyInGo/core/pipeline"
	"github.com/obgnail/ScrapyInGo/core/scheduler"
	"github.com/obgnail/ScrapyInGo/core/spider"
)

const (
	ReqChanMaxLen  = 1024
	RespChanMaxLen = 1024
	ItemChanMaxLen = 2048

	DefaultSemaphore = 2

	SleepTimeWhenHasNoRequest = 100 * time.Millisecond
)

var ParseFunc func(r *entity.Response) (interface{}, error)

type Engine struct {
	spider      spider.Spider
	scheduler   scheduler.Scheduler
	downloader  downloader.Downloader
	pipeline    pipeline.Pipeline
	middlewares []middleware.Middleware

	reqChan   chan *entity.Request
	respChan  chan *entity.Response
	itemChan  chan interface{}
	closeChan chan struct{}

	semaphore uint
}

func New(
	spider spider.Spider,
	scheduler scheduler.Scheduler,
	downloader downloader.Downloader,
	pipeline pipeline.Pipeline,
	middleware []middleware.Middleware,
) *Engine {
	e := &Engine{
		spider:      spider,
		scheduler:   scheduler,
		downloader:  downloader,
		pipeline:    pipeline,
		middlewares: middleware,
		reqChan:     make(chan *entity.Request, ReqChanMaxLen),
		respChan:    make(chan *entity.Response, RespChanMaxLen),
		itemChan:    make(chan interface{}, ItemChanMaxLen),
		closeChan:   make(chan struct{}, 1),
	}
	return e
}

func Default(sp spider.Spider) *Engine {
	return New(
		sp,
		scheduler.NewSimpleScheduler(),
		downloader.NewDefaultDownloader(),
		pipeline.NewDefaultPipeline(),
		[]middleware.Middleware{middleware.NewDefaultMiddleware()},
	)
}

func (e *Engine) PushRequest(req *entity.Request) {
	if e.scheduler == nil {
		log.Fatal("spider has no scheduler")
	}
	go func() {
		e.scheduler.Push(req)
	}()
}

func (e *Engine) PushRequests(requests []*entity.Request) {
	if e.scheduler == nil {
		log.Fatal("spider has no scheduler")
	}
	go func() {
		for _, req := range requests {
			e.scheduler.Push(req)
		}
	}()
}

func (e *Engine) SetThreadLimit(num uint) {
	e.semaphore = num
}

func (e *Engine) SetScheduler(scheduler scheduler.Scheduler) {
	e.scheduler = scheduler
}

func (e *Engine) SetDownloader(downloader downloader.Downloader) {
	e.downloader = downloader
}

func (e *Engine) SetPipeline(pipeline pipeline.Pipeline) {
	e.pipeline = pipeline
}

func (e *Engine) AppendMiddleware(middleware middleware.Middleware) {
	e.middlewares = append(e.middlewares, middleware)
}

func (e *Engine) GetScheduler() scheduler.Scheduler {
	return e.scheduler
}

func (e *Engine) GetDownloader() downloader.Downloader {
	return e.downloader
}

func (e *Engine) GetPipeline() pipeline.Pipeline {
	return e.pipeline
}

func (e *Engine) GetMiddlewares() []middleware.Middleware {
	return e.middlewares
}

func (e *Engine) middlewareProcessRequest(req *entity.Request) *entity.Request {
	for _, m := range e.middlewares {
		req = m.ProcessRequest(req, e.spider)
	}
	return req
}

func (e *Engine) middlewareProcessResponse(resp *entity.Response) *entity.Response {
	for _, m := range e.middlewares {
		resp = m.ProcessResponse(resp, e.spider)
	}
	return resp
}

func (e *Engine) middlewareProcessItem(item interface{}) interface{} {
	for _, m := range e.middlewares {
		item = m.ProcessItem(item, e.spider)
	}
	return item
}

func (e *Engine) middlewareSpiderOpened() {
	for _, m := range e.middlewares {
		m.SpiderOpened(e.spider)
	}
}

func (e *Engine) startScheduler() {
	if e.scheduler == nil {
		log.Fatal("engine has no scheduler")
	}
	for {
		if e.scheduler.Len() != 0 {
			e.reqChan <- e.scheduler.Pop()
		} else {
			time.Sleep(SleepTimeWhenHasNoRequest)
		}
	}
}

func (e *Engine) startDownloader() {
	if e.downloader == nil {
		log.Fatal("engine has no downloader")
	}
	semaphoreChan := make(chan struct{}, e.semaphore)
	for {
		select {
		case req := <-e.reqChan:
			req = e.middlewareProcessRequest(req)
			// 此请求被放弃
			if req == nil {
				continue
			}
			semaphoreChan <- struct{}{}
			go func() {
				defer func() { <-semaphoreChan }()
				log.Printf("[INFO] scraping url: %s\n", req.GetUrl())
				resp, err := e.downloader.Download(req)
				if err != nil {
					log.Printf("[WARN] scraping error url: %s\n%s\n", req.GetUrl(), errors.Trace(err))
					if resp != nil {
						resp.GetRespObj().Body.Close()
					}
					if req.Errback != nil {
						req.Errback(req, err)
					}
					return
				}
				if resp.GetRespObj() == nil {
					return
				}
				e.respChan <- resp
			}()
		}
	}
}

func (e *Engine) startParse() {
	if e.spider == nil {
		log.Fatal("engine has no spider")
	}
	for {
		select {
		case resp := <-e.respChan:
			resp = e.middlewareProcessResponse(resp)
			// 此响应被放弃
			if resp == nil {
				continue
			}
			go func() {
				if resp.Callback == nil {
					ParseFunc = e.spider.Parse
				} else {
					ParseFunc = resp.Callback
				}
				item, err := ParseFunc(resp)
				if resp != nil {
					resp.GetRespObj().Body.Close()
				}
				if err != nil {
					log.Printf("[WARN] parse error: url: %s\n%s\n", resp.GetRequest().GetUrl(), errors.Trace(err))
					if resp.Errback != nil {
						resp.Errback(resp.Request, err)
					}
					return
				}
				if item == nil {
					return
				}
				e.itemChan <- item
			}()
		}
	}
}

func (e *Engine) startPipeline() {
	if e.pipeline == nil {
		log.Fatal("engine has no pipeline")
	}
	for {
		select {
		case item := <-e.itemChan:
			item = e.middlewareProcessItem(item)
			if item == nil {
				continue
			}
			go func() {
				if err := e.pipeline.ProcessItem(item, e.spider); err != nil {
					log.Printf("processItem err %s\n", errors.Trace(err))
				}
			}()
		}
	}
}

func (e *Engine) Crawl() {
	e.middlewareSpiderOpened()
	go e.startPipeline()
	go e.startParse()
	go e.startDownloader()
	go e.startScheduler()
	<-e.closeChan
}
