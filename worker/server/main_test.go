package main

import (
	"GoP2PSpider/config"
	"GoP2PSpider/rpcsupport"
	"GoP2PSpider/types"
	"testing"
)

func TestWorkerCallEngine(t *testing.T) {
	const engineHost = ":9000"
	client, _ := rpcsupport.NewClient(engineHost)
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
	client.Call(config.EngineDataReceiver, torrent, "")
}
