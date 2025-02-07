package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"
)

var DurationSec int

func main() {
	var (
		network     = flag.String("network", "udp4", "network")
		addr        = flag.String("addr", "", "example: 1.2.3.4:80")
		sni         = flag.String("sni", "", "domain,empty is skip")
		quicVersion = flag.String("quic-version", "43", "support 39,43,44")
		local       = flag.String("bind", "", "bind local ip")
		name        = flag.String("file", "test.flv", "specify i/o flv file")
		buffer      = flag.Int("buffer", 102400, "buffer size in byte")
		rtmpType    = flag.Bool("type", true, "whether to pull or publish,true is pull")
		skip        = flag.Bool("skip", true, "whether a client verifies the server's certificate chain and host name")
		client      = flag.Int("client", 1, "paral client number")
		reqcount    = flag.Int("reqcount", 0, "total request count")
		interval    = flag.Int("interval", 50, "interval duration(micro-second) between 2 requests")
		duration    = flag.Int("duration", 86400, "duration time(second) for each-requests")
	)

	flag.Parse()
	DurationSec = *duration

	rawurl := os.Args[len(os.Args)-1]

	//filename := time.Now().Format("2006-01-02-15-04-05-999-") + *name
	filename := ""
	if *client == 1 {
		filename = time.Now().Format("2006-01-02-15-04-") + *name
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		fmt.Println("url parsed error:", rawurl, err)
		return
	}

	if *sni == "" {
		*sni = u.Host
	}

	tlsCfg, cfg := parseCfg(*quicVersion, *sni, *skip)

	buffer_size := *buffer

	quit := make(chan int, *client)

	idx := 0
	for idx < *client {
		go run(idx,
			*network,
			*local,
			*addr,
			rawurl,
			filename,
			tlsCfg,
			cfg,
			buffer_size,
			u,
			*rtmpType,
			quit)
		idx += 1
		time.Sleep(time.Duration(*interval) * time.Millisecond)
	}

	max := *client + *reqcount
	idx = 0
	for idx < max {
		<-quit
		idx += 1
		if *reqcount > 0 {
			go run(idx+*client,
				*network,
				*local,
				*addr,
				rawurl,
				filename,
				tlsCfg,
				cfg,
				buffer_size,
				u,
				*rtmpType,
				quit)
		}
	}
}
