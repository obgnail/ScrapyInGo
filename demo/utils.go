package demo

import (
	"github.com/juju/errors"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"log"
	"net/http"
	"os"
)

func getContentFromResponse(resp *entity.Response) ([]byte, error) {
	if resp.GetStatus() != 200 {
		return nil, errors.Errorf("status is not 200:[%s]", resp.Request.GetUrl())
	}
	return resp.GetContent()
}

func newSimpleRequest(
	url string,
	callback entity.CallbackFunc,
	errBack entity.ErrbackFunc,
	meta map[string]interface{},
) (*entity.Request, error) {
	reqObj, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req := entity.NewRequest(reqObj, 100, false, callback, errBack, meta)
	return req, nil
}

func ValidName(name []byte) []byte {
	return ValidNameRegexp.ReplaceAll(name, []byte("_"))
}

func MdirIfNotExist(dirPath string) {
	_, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			if e := os.Mkdir(dirPath, 0755); e != nil {
				log.Printf("[ERROR] mkdir:%s", dirPath)
			}
		}
	}
}
