package demo

import (
	"github.com/obgnail/ScrapyInGo/core/spider"
	"io/ioutil"
	"log"
	"path/filepath"
)

const (
	BasePath = "d:\\tmp\\download"
)

type StoreMangaPipeline struct{}

func (p *StoreMangaPipeline) ProcessItem(item interface{}, sp spider.Spider) error {
	manga := item.(*Manga)
	dirPath := filepath.Join(BasePath, manga.dirName)
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

func NewStoreMangaPipeline() *StoreMangaPipeline {
	return &StoreMangaPipeline{}
}
