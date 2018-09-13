package main

import (
	"GoP2PSpider/engine/server/service"
	"GoP2PSpider/rpcsupport"
	"flag"
	"fmt"
	"log"
)

func main() {
	dataServiceHost := flag.String("data_service_host", "", "data_service_host")
	enginePort := flag.Int("port", 0, "the port listen on localhost, waiting for worker call")
	flag.Parse()

	if *dataServiceHost == "" {
		fmt.Println("must specify a dataServiceHost")
		return
	}
	if *enginePort == 0 {
		fmt.Println("must specify a engine port")
		return
	}

	client, e := rpcsupport.NewClient(*dataServiceHost)
	if e != nil {
		panic(e)
	}
	log.Fatal(rpcsupport.ServeRpc(fmt.Sprintf(":%d", *enginePort), &service.DataReceiver{
		Client: client,
	}))
}
