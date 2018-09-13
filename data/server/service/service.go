package service

import (
	"GoP2PSpider/config"
	"GoP2PSpider/types"
	"context"
	"gopkg.in/olivere/elastic.v5"
	"log"
)

type DataService struct {
	Client *elastic.Client
}

func (d *DataService) Save(torrent types.Torrent, result *string) error {
	e := Save(d.Client, torrent)
	if e == nil {
		*result = "ok"
		log.Printf("Success saving %v", torrent)
	} else {
		*result = "fail"
		log.Printf("Error saving %v, %v", torrent, e)
	}
	return e
}

func Save(client *elastic.Client, torrent types.Torrent) error {
	_, e := client.Index().
		Index(config.ElasticIndex).
		Type(config.ElasticType).
		Id(torrent.InfoHashHex).
		BodyJson(torrent).
		Do(context.Background())
	if e != nil {
		log.Printf("es create index fail %v", e)
		return e
	}
	return nil
}
