package main

import "github.com/obgnail/ScrapyInGo/demo"

const (
	spiderName    = "nhentai"
	storeBasePath = "y:\\comic2"
	proxyUrl      = "http://127.0.0.1:7890"
	threadNum     = 2

	UA     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	Cookie = "cf_clearance=daa9d77091a0e1b953eaeff3cf032ad0d014e6f4-1626532362-0-150; csrftoken=Zxvfpgcdhf1uDMzzlwvsB4gC7TICWi6TRZcCubCnfEnECwY8MzfVbXgg6Bzw2Ytf; sessionid=uzckbjtbrie0smci1z0z5y13de5nf59g"
)

func main() {
	spider := demo.New(
		spiderName,
		threadNum,
		map[string]string{"cookie": Cookie, "user-agent": UA},
		proxyUrl,
		storeBasePath,
	)
	spider.QueueStartUrl()
	spider.Crawl()
}
