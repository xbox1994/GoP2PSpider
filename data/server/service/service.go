package service

import (
	"GoP2PSpider/config"
	"GoP2PSpider/types"
	"context"
	"gopkg.in/olivere/elastic.v5"
	"log"
	"reflect"
)

type DataService struct {
	Client *elastic.Client
}

func (d *DataService) Save(torrent *types.Meta, result *string) error {
	query := elastic.NewTermQuery("_id", torrent.Hash)
	searchResult, e := d.Client.Search().
		Index(config.ElasticIndex).
		Type(config.ElasticType).
		Query(query).
		Do(context.Background())
	if searchResult != nil && searchResult.Hits.TotalHits > 0 {
		log.Printf("Torrent existed, won't be save: %s", torrent)
		return e
	}

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

var pageSize = 10

func (d *DataService) Query(param *types.QueryParam, result *types.QueryResult) error {
	log.Printf("Query: %v", param)
	searchResult, e := d.Client.
		Search(config.ElasticIndex).
		Type(config.ElasticType).
		Query(elastic.NewQueryStringQuery(param.Q)).
		Size(pageSize).
		From(param.Start).
		Do(context.Background())
	if e != nil {
		log.Printf("Error query: %v, error: %v", param, e)
	}

	result.Query = param.Q
	result.Hits = searchResult.TotalHits()
	result.Start = param.Start
	result.Items = searchResult.Each(reflect.TypeOf(types.Meta{}))
	result.PrevStart = param.Start - pageSize
	result.NextStart = result.Start + len(result.Items)
	return nil
}

func Save(client *elastic.Client, torrent types.Meta) error {
	_, e := client.Index().
		Index(config.ElasticIndex).
		Type(config.ElasticType).
		Id(torrent.Hash).
		BodyJson(torrent).
		Id(torrent.Hash).
		Do(context.Background())
	if e != nil {
		log.Printf("es create index fail %v", e)
		return e
	}
	return nil
}
