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
	esHost := flag.String("esHost", "http://127.0.0.1:9200", "the port listen on localhost, waiting for engine call")
	port := flag.Int("p", 9000, "the port listen on localhost, waiting for engine call")
	flag.Parse()

	client, e := elastic.NewClient(elastic.SetURL(*esHost), elastic.SetSniff(false))
	if e != nil {
		panic(e)
	}
	fmt.Println("data service starting up finished, waiting for request...")
	log.Fatal(rpcsupport.ServeRpc(fmt.Sprintf(":%d", *port), &service.DataService{
		Client: client,
	}))
}
