package entity

import (
	"github.com/juju/errors"
	"io/ioutil"
	"net/http"
)

type Response struct {
	RaspObj  *http.Response
	Request  *Request
	Callback CallbackFunc
	Errback  ErrbackFunc
	Meta     map[string]interface{}
}

func NewResponse(
	response *http.Response,
	callback CallbackFunc,
	errback ErrbackFunc,
	meta map[string]interface{},
) *Response {
	return &Response{
		RaspObj:  response,
		Request:  nil,
		Callback: callback,
		Errback:  errback,
		Meta:     meta,
	}
}

func FromRequest(request *Request) *Response {
	r := &Response{}
	r.Request = request
	r.Callback = request.Callback
	r.Errback = request.Errback
	r.Meta = request.Meta
	return r
}

func (r *Response) GetRespObj() *http.Response {
	return r.RaspObj
}

func (r *Response) SetRespObj(raspObj *http.Response) {
	r.RaspObj = raspObj
}

func (r *Response) GetRequest() *Request {
	return r.Request
}

func (r *Response) GetContent() ([]byte, error) {
	resp, err := ioutil.ReadAll(r.GetRespObj().Body)
	if err != nil {
		return nil, errors.Errorf("get content error:[%s]", err)
	}
	return resp, nil
}

func (r *Response) GetStatus() int {
	return r.GetRespObj().StatusCode
}
