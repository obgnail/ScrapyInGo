package entity

import (
	"github.com/juju/errors"
	"net/http"
)

//type Context struct {
//	*Request
//	*Response
//}

func ProcessRequest(client *http.Client, req *Request) (*Response, error) {
	respObj, err := client.Do(req.GetReqObj())
	if err != nil {
		return nil, errors.Errorf("download failed, err:[%s]", err.Error())
	}
	resp := FromRequest(req)
	resp.SetRespObj(respObj)
	req.SetResponse(resp)
	return resp, nil
}
