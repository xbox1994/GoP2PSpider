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

func (d *DataService) Save(torrent *types.Torrent, result *string) error {
	query := elastic.NewTermQuery("_id", torrent.InfoHashHex)
	searchResult, e := d.Client.Search().
		Index(config.ElasticIndex).
		Type(config.ElasticType).
		Query(query).
		Do(context.Background())
	if searchResult != nil && searchResult.Hits.TotalHits > 0 {
		log.Printf("Torrent existed, won't be save: %s", torrent)
		return e
	}

	log.Printf("Torrent received in data service, will be save to es: %s", torrent)
	saveErr := Save(d.Client, *torrent)
	if saveErr == nil {
		*result = "ok"
		log.Printf("Success saving %s", torrent)
	} else {
		*result = "fail"
		log.Printf("Error saving %s, %v", torrent, saveErr)
	}
	return saveErr
}

func Save(client *elastic.Client, torrent types.Torrent) error {
	_, e := client.Index().
		Index(config.ElasticIndex).
		Type(config.ElasticType).
		Id(torrent.InfoHashHex).
		BodyJson(torrent).
		Id(torrent.InfoHashHex).
		Do(context.Background())
	if e != nil {
		log.Printf("es create index fail %v", e)
		return e
	}
	return nil
}
