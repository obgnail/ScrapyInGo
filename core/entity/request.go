package entity

import (
	"net/http"
)

type CallbackFunc func(r *Response) (interface{}, error)
type ErrbackFunc func(r *Request, err error)

type Request struct {
	ReqObj     *http.Request
	response   *Response
	priority   uint
	retries    uint
	dontFilter bool
	Callback   CallbackFunc // 请求成功的时候回调此函数
	Errback    ErrbackFunc  // 请求或解析失败的时候回调此函数
	Meta       map[string]interface{}
}

func NewRequest(
	reqObj *http.Request,
	priority uint,
	dontFilter bool,
	callback CallbackFunc,
	errback ErrbackFunc,
	Meta map[string]interface{},
) *Request {
	return &Request{
		ReqObj:     reqObj,
		response:   nil,
		priority:   priority,
		retries:    0,
		dontFilter: dontFilter,
		Callback:   callback,
		Errback:    errback,
		Meta:       Meta,
	}
}

func (r *Request) GetReqObj() *http.Request {
	return r.ReqObj
}

func (r *Request) GetUrl() string {
	return r.ReqObj.URL.String()
}

func (r *Request) GetPriority() uint {
	return r.priority
}

func (r *Request) GetResponse() *Response {
	return r.response
}

func (r *Request) SetResponse(resp *Response) {
	r.response = resp
}

func (r *Request) GetReTries() uint {
	return r.retries
}

func (r *Request) IncReTries() {
	r.retries++
}
