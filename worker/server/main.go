package main

import (
	"GoP2PSpider/config"
	"GoP2PSpider/rpcsupport"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/neoql/btlet"
)

func main() {
	dataHost := flag.String("data_host", "0.0.0.0:9000", "data service receive host")
	client, e := rpcsupport.NewClient(*dataHost)
	if e != nil {
		panic(e)
	}
	flag.Parse()

	builder := btlet.NewSnifferBuilder()
	p := btlet.NewSimplePipelineWithBuf(512)
	s := builder.NewSniffer(p)
	go s.Sniff(context.TODO())

	total := 0
	go statistic(&total)
	fmt.Println("Start crawl ...")

	for meta := range p.MetaChan() {
		meta.Hash = fmt.Sprintf("%x", meta.Hash)
		log.Printf("metadata: %v", meta)

		result := ""
		e := client.Call(config.DataService, meta, &result)
		if e != nil {
			log.Printf("worker call data error: %v", e)
		}

		total++
		os.Stdout.WriteString(fmt.Sprintf("\rHave already sniff %d torrents.", total))
	}
}

func statistic(total *int) {
	last := 0
	for range time.Tick(time.Minute) {
		t := *total
		sub := t - last
		last = t
		fmt.Printf("\rSniffed %d torrents last minute.\n", sub)
	}
}
