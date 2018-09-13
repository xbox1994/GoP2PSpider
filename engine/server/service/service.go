package service

import (
	"GoP2PSpider/config"
	"GoP2PSpider/types"
	"log"
	"net/rpc"
)

type DataReceiver struct {
	Client *rpc.Client
}

func (d *DataReceiver) Receive(torrent *types.Torrent, result *string) error {
	go func() {
		log.Printf("Torrent received in engine, will be sent to data service: \n%s", torrent)
		d.Client.Call(config.DataService, torrent, "")
	}()
	return nil
}
