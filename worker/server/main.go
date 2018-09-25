package main

import (
	"GoP2PSpider/config"
	"GoP2PSpider/rpcsupport"
	"GoP2PSpider/types"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/neoql/btlet"
	"github.com/neoql/btlet/bt"
)

func main() {
	dataHost := flag.String("data_host", "0.0.0.0:9000", "data service receive host")
	client, e := rpcsupport.NewClient(*dataHost)
	if e != nil {
		panic(e)
	}
	flag.Parse()

	builder := btlet.NewSnifferBuilder()
	// 如果想要限制性能可以通过设置builder.MaxWorkers来设置。数值可以根据情况设置
	// example:
	// builder.MaxWorkers = 256
	p := NewSimplePipelineWithBuf(512)
	s := builder.NewSniffer(p)
	go s.Sniff(context.TODO())

	total := 0
	fmt.Println("Start crawl ...")
	for mo := range p.MetaChan() {

		result := ""
		files := make([]types.File, len(mo.files))
		for _, file := range mo.files {
			if file.size != 0 {
				files = append(files, types.File{
					Path: file.path,
					Size: file.size,
				})
			}
		}
		e := client.Call(config.DataServiceSave, &types.Meta{
			Hash:  mo.link,
			Name:  mo.title,
			Size:  mo.size,
			Files: files,
		}, &result)
		if e != nil {
			log.Printf("worker call data error: %v", e)
		}

		total++
		os.Stdout.WriteString("\r")
		fmt.Println("-------------------------------------------------------------")
		fmt.Println(mo.link)
		fmt.Println("Title:", mo.title)
		fmt.Println("Size:", mo.size)
		for _, f := range mo.files {
			fmt.Println("-", f.path)
		}
		os.Stdout.WriteString(fmt.Sprintf("Have already sniff %d torrents.", total))
	}
}

// SimplePipeline is a simple pipeline
type SimplePipeline struct {
	ch chan metaOutline
}

// NewSimplePipeline will returns a simple pipline
func NewSimplePipeline() *SimplePipeline {
	return &SimplePipeline{
		ch: make(chan metaOutline),
	}
}

// NewSimplePipelineWithBuf will returns a simple pipline with buffer
func NewSimplePipelineWithBuf(bufSize int) *SimplePipeline {
	return &SimplePipeline{
		ch: make(chan metaOutline, bufSize),
	}
}

// DisposeMeta will handle meta
func (p *SimplePipeline) DisposeMeta(hash string, meta bt.RawMeta) {
	var mo metaOutline
	err := meta.FillOutline(&mo)
	if err != nil {
		return
	}
	mo.SetHash(hash)
	p.ch <- mo
}

// MetaChan returns a Meta channel
func (p *SimplePipeline) MetaChan() <-chan metaOutline {
	return p.ch
}

// PullTrackerList always return nil, false
func (p *SimplePipeline) PullTrackerList(string) ([]string, bool) {
	return nil, false
}

type metaOutline struct {
	link  string
	title string
	size  uint64
	files []file
}

type file struct {
	path string
	size uint64
}

func (mo *metaOutline) SetName(name string) {
	mo.title = name
}

func (mo *metaOutline) SetHash(hash string) {
	mo.link = fmt.Sprintf("%x", hash)
}

func (mo *metaOutline) AddFile(path string, size uint64) {
	mo.files = append(mo.files, file{path, size})
	mo.size += size
}
