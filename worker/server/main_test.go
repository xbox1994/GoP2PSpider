package main

import (
	"GoP2PSpider/config"
	"GoP2PSpider/rpcsupport"
	"GoP2PSpider/types"
	"testing"
	"time"
)

func TestWorkerCallEngine(t *testing.T) {
	const dataHost = ":9000"
	client, _ := rpcsupport.NewClient(dataHost)
	tFile := types.TFile{
		Name:   "tfboy",
		Length: 100,
	}
	torrent := types.Torrent{
		InfoHashHex: "hash",
		Name:        "zhongziname",
		Length:      3,
		Files:       []*types.TFile{&tFile},
	}
	for {
		time.Sleep(time.Second)
		result := ""
		e := client.Call(config.DataService, torrent, &result)
		if e != nil {
			panic(e)
		}
	}
}
