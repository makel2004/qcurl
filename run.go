package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/lucas-clemente/quic-go"
)

func run(idx int,
	network, local, addr, rawurl, filename string,
	tlsCfg *tls.Config, cfg *quic.Config,
	buffer_size int,
	//buffer []byte,
	//dst *os.File,
	u *url.URL, rtmpType bool,
	quit chan int) {
	defer func() {
		fmt.Println("client exit:", idx)
		quit <- idx
	}()

	if addr == "" {
		HostSplit := strings.Split(u.Host, ":")
		var port string
		if len(HostSplit) >= 2 {
			port = HostSplit[1]
		} else {
			switch u.Scheme {
			case "http":
				port = "80"
			case "https":
				port = "443"
			case "rtmp":
				port = "1935"
			default:
			}
		}

		ips, err := net.LookupIP(HostSplit[0])
		if err != nil || len(ips) == 0 {
			fmt.Println(err)
			return
		}
		printDNS(u.Host, ips)

		addr = ips[0].String() + ":" + port
	}
	fmt.Printf("dial:%+v\n", addr)
	buffer := make([]byte, buffer_size)

	local_filename := fmt.Sprintf("%d-%s", idx, filename)
	dst_file, err := os.OpenFile(local_filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		dst_file.Sync()

		stat, _ := dst_file.Stat()
		if stat.Size() == 0 {
			os.Remove(dst_file.Name())
		}

		dst_file.Close()
	}()

	switch u.Scheme {
	case "http":
		h1OverQUIC(idx, network, local, addr, rawurl, tlsCfg, cfg, buffer, dst_file)
	case "https":
		h2OverQUIC(idx, network, local, addr, rawurl, tlsCfg, cfg, buffer, dst_file)
	case "rtmp":
		rtmpOverQUIC(idx, network, local, addr, rawurl, tlsCfg, cfg, filename, rtmpType)

	default:
		fmt.Println("unsupport")
	}
}
