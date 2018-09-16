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
	log.Printf("Torrent received in engine, will be sent to data service: \n%s", torrent)
	r := ""
	e := d.Client.Call(config.DataService, *torrent, &r)
	if e == nil {
		*result = "ok"
		log.Printf("Success saving %s", torrent)
	} else {
		*result = "fail"
		log.Printf("Error saving %s, %v", torrent, e)
	}
	return e
}
