package entity

import (
	"net/http"
)

type CallbackFunc func(r *Response) (interface{}, error)
type ErrbackFunc func(r *Response, err error)

type Request struct {
	ReqObj     *http.Request
	priority   uint
	dontFilter bool
	Callback   CallbackFunc
	Errback    ErrbackFunc
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
		priority:   priority,
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
