package main

import (
	"GoP2PSpider/config"
	"GoP2PSpider/rpcsupport"
	"GoP2PSpider/types"
	"GoP2PSpider/util/bencode"
	"GoP2PSpider/util/bep"
	"GoP2PSpider/util/dht"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	Dir = "torrents"
)

func home() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env)
}

func newTorrent(meta []byte, infohashHex string) (*types.Torrent, error) {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return nil, err
	}
	t := &types.Torrent{InfoHashHex: infohashHex}
	if name, ok := dict["name.utf-8"].(string); ok {
		t.Name = name
	} else if name, ok := dict["name"].(string); ok {
		t.Name = name
	}
	if length, ok := dict["length"].(int64); ok {
		t.Length = length
	}
	var total int64
	if files, ok := dict["files"].([]interface{}); ok {
		for _, file := range files {
			var filename string
			var fileLength int64
			if f, ok := file.(map[string]interface{}); ok {
				if inter, ok := f["path.utf-8"].([]interface{}); ok {
					path := make([]string, len(inter))
					for i, v := range inter {
						path[i] = fmt.Sprint(v)
					}
					filename = strings.Join(path, "/")
				} else if inter, ok := f["path"].([]interface{}); ok {
					path := make([]string, len(inter))
					for i, v := range inter {
						path[i] = fmt.Sprint(v)
					}
					filename = strings.Join(path, "/")
				}
				if length, ok := f["length"].(int64); ok {
					fileLength = length
					total += fileLength
				}
				t.Files = append(t.Files, &types.TFile{Name: filename, Length: fileLength})
			}
		}
	}
	if t.Length == 0 {
		t.Length = total
	}
	if len(t.Files) == 0 {
		t.Files = append(t.Files, &types.TFile{Name: t.Name, Length: t.Length})
	}
	return t, nil
}

type blacklist struct {
	container    sync.Map
	expiredAfter time.Duration
}

func newBlackList(expiredAfter time.Duration) *blacklist {
	b := &blacklist{expiredAfter: expiredAfter}
	go b.sweep()
	return b
}

func (b *blacklist) in(addr *net.TCPAddr) bool {
	key := addr.String()
	v, ok := b.container.Load(key)
	if !ok {
		return false
	}
	c := v.(time.Time)
	if c.Sub(time.Now()) > b.expiredAfter {
		b.container.Delete(key)
		return false
	}
	return true
}

func (b *blacklist) add(addr *net.TCPAddr) {
	b.container.Store(addr.String(), time.Now())
}

func (b *blacklist) sweep() {
	for range time.Tick(10 * time.Second) {
		now := time.Now()
		b.container.Range(func(k, v interface{}) bool {
			c := v.(time.Time)
			if c.Sub(now) > b.expiredAfter {
				b.container.Delete(k)
			}
			return true
		})
	}
}

type p2pspider struct {
	laddr        string
	maxFriends   int
	maxPeers     int
	secret       string
	timeout      time.Duration
	blacklist    *blacklist
	dir          string
	engineClient *rpc.Client
}

func (p *p2pspider) run() {
	tokens := make(chan struct{}, p.maxPeers)
	dht, err := dht.New(
		p.laddr,
		dht.MaxFriendsPerSec(p.maxFriends),
		dht.Secret(p.secret),
	)
	if err != nil {
		panic(err)
	}
	log.Println("running, wait a few minutes...")
	for ac := range dht.Announce {
		tokens <- struct{}{}
		go p.work(ac, tokens)
	}
}

func (p *p2pspider) work(ac *dht.Announce, tokens chan struct{}) {
	defer func() {
		<-tokens
	}()
	if p.isExist(ac.InfohashHex) {
		return
	}
	if p.blacklist.in(ac.Peer) {
		return
	}
	peer := bep.New(
		string(ac.Infohash),
		ac.Peer.String(),
		bep.Timeout(p.timeout),
	)
	data, err := peer.Fetch()
	if err != nil {
		p.blacklist.add(ac.Peer)
		return
	}
	_, err = p.save(ac.InfohashHex, data)
	if err != nil {
		return
	}
	t, err := newTorrent(data, ac.InfohashHex)
	if err != nil {
		return
	}
	log.Println(t)

	// call engine client to send t
	p.engineClient.Call(config.EngineDataReceiver, t, "")
}

func (p *p2pspider) isExist(infohashHex string) bool {
	name, _ := p.pathname(infohashHex)
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (p *p2pspider) save(infohashHex string, data []byte) (string, error) {
	name, dir := p.pathname(infohashHex)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return "", err
	}
	defer f.Close()
	d, err := bencode.Decode(bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	_, err = f.Write(bencode.Encode(map[string]interface{}{
		"info": d,
	}))
	if err != nil {
		return "", err
	}
	return name, nil
}

func (p *p2pspider) pathname(infohashHex string) (name string, dir string) {
	dir = path.Join(p.dir, infohashHex[:2], infohashHex[len(infohashHex)-2:])
	name = path.Join(dir, infohashHex+".torrent")
	return
}

func main() {
	addr := flag.String("a", "0.0.0.0", "listen on given address")
	port := flag.Int("p", 6881, "worker listen on given port")
	maxFriends := flag.Int("f", 500, "max friends to make with per second")
	peers := flag.Int("e", 400, "max peers(TCP) to connenct to download torrent file")
	timeout := flag.Duration("t", 10*time.Second, "max time allowed for downloading torrent file")
	secret := flag.String("s", "$p2pspider$", "token secret")
	dir := flag.String("d", path.Join(home(), Dir), "the directory to store the torrent file")
	verbose := flag.Bool("v", true, "run in verbose mode")
	engineHost := flag.String("eh", "0.0.0.0:9000", "engine data receive host")
	flag.Parse()
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		panic(err)
	}
	if *verbose {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	client, e := rpcsupport.NewClient(*engineHost)
	if e != nil {
		panic(e)
	}
	p := &p2pspider{
		laddr:        fmt.Sprintf("%s:%d", *addr, *port),
		timeout:      *timeout,
		maxFriends:   *maxFriends,
		maxPeers:     *peers,
		secret:       *secret,
		dir:          absDir,
		blacklist:    newBlackList(10 * time.Minute),
		engineClient: client,
	}

	p.run()
}