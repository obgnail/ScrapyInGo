package demo

import (
	"github.com/obgnail/ScrapyInGo/core/spider"
	"io/ioutil"
	"log"
	"path/filepath"
)

type StoreMangaPipeline struct {
	storeBasePath string
}

func (p *StoreMangaPipeline) ProcessItem(item interface{}, sp spider.Spider) error {
	manga := item.(*Manga)
	dirPath := filepath.Join(p.storeBasePath, manga.dirName)
	MdirIfNotExist(dirPath)
	filePath := filepath.Join(dirPath, manga.fileName)
	err := ioutil.WriteFile(filePath, manga.content, 0666)
	if err != nil {
		log.Printf("[ERROR] save successful url:%s\n", filePath)
		return nil
	}
	log.Printf("[INFO] save manga successful url:%s\n", filePath)
	return nil
}

func NewStoreMangaPipeline(basePath string) *StoreMangaPipeline {
	return &StoreMangaPipeline{basePath}
}
