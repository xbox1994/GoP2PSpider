package main

import (
	"GoP2PSpider/data/server/service"
	"GoP2PSpider/rpcsupport"
	"flag"
	"fmt"
	"gopkg.in/olivere/elastic.v5"
	"log"
)

func main() {
	port := flag.Int("port", 0, "the port listen on localhost, waiting for engine call")
	flag.Parse()
	if *port == 0 {
		fmt.Println("must specify a data service port")
		return
	}

	client, e := elastic.NewClient(elastic.SetSniff(false))
	if e != nil {
		panic(e)
	}
	fmt.Println("data service starting up finished, waiting for request...")
	log.Fatal(rpcsupport.ServeRpc(fmt.Sprintf(":%d", *port), &service.DataService{
		Client: client,
	}))
}
