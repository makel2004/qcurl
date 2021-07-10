package main

import (
    "crypto/tls"
    "fmt"
    "io"
    "net/http"
    "net/http/httputil"
    "time"

    quic "github.com/lucas-clemente/quic-go"
    "github.com/lucas-clemente/quic-go/h2quic"
)

func h2OverQUIC(idx int,
    network, local, addr, rawurl string,
    tlsCfg *tls.Config, cfg *quic.Config,
    buffer []byte, dst io.Writer) {

    //fmt.Printf("%+v -> %+v\n", local, addr)
    roundTripper := &h2quic.RoundTripper{
        TLSClientConfig: tlsCfg,
        QuicConfig:      cfg,
        Dial:            dialFunc(local, addr),
    }

    defer roundTripper.Close()

    client := &http.Client{
        Transport: roundTripper,
    }

    dial := time.Now()
    req, err := http.NewRequest("GET", rawurl, nil)
    if err != nil {
        fmt.Println(err)

        return
    }

    data, err := httputil.DumpRequestOut(req, true)
    if err != nil {
        fmt.Println(err)

        return
    }

    fmt.Printf("client:%v %v\n", idx, string(data))

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err)

        return
    }

    readresp := time.Now()
    fmt.Printf("client:%v recv resp after %+v\n", idx, readresp.Sub(dial).String())

    rp, err := DumpResponse(resp)
    fmt.Println(string(rp))
    if err != nil {
        fmt.Println(err)
    }

    for {
        if _, err := io.CopyN(dst, resp.Body, 4096); err != nil {
            fmt.Println("80:", err)
            break
        }
        cn := time.Now().Sub(readresp).Seconds()
        if cn > 10 {
            fmt.Printf("client:%v close after %v seconds\n", idx, cn)
            break
        }
    }

    if resp != nil && resp.Body != nil {
        resp.Body.Close()
    }
}
