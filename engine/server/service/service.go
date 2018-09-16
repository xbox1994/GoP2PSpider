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

func (d *DataReceiver) Receive(torrent types.Torrent, result *string) error {
	e := d.Client.Call(config.DataService, torrent, &result)
	if e == nil {
		*result = "ok"
		log.Printf("Success saving %v", torrent)
	} else {
		*result = "fail"
		log.Printf("Error saving %v, %v", torrent, e)
	}
	return e
}
