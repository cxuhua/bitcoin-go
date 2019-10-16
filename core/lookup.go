package core

import (
	"bitcoin/config"
	"context"
	"log"
	"net"
	"sync"
	"time"
)

var (
	MWG = sync.WaitGroup{}
)

//run lookup
func lookupseeds(ctx context.Context, conf *config.Config) []IPPort {
	//lookip
	ips := []IPPort{}
	for _, v := range conf.Seeds {
		if ctx.Err() != nil {
			return ips
		}
		lips, err := net.LookupIP(v)
		if err != nil {
			log.Println("LOOKUP ip error ", v, err)
			continue
		}
		for _, v := range lips {
			ips = append(ips, IPPort{ip: v, port: conf.ListenPort})
		}
		log.Println("LOOKUP", v, "Count=", len(ips))
	}
	return ips
}

func getconnip(ips []IPPort, idx int, conf *config.Config) (IPPort, bool) {
	local := IPPort{
		ip:   net.ParseIP(conf.LocalIP),
		port: conf.ListenPort,
	}
	if idx < 0 || idx >= len(ips) {
		return local, false
	}
	ip := ips[idx]
	if ip.Equal(local) {
		return local, false
	}
	if !ip.IsEnable() {
		return local, false
	}
	if !Addrs.IsConnect(ip) {
		return local, false
	}
	return ip, true
}

//启动
func StartLookUp(ctx context.Context) {
	defer MWG.Done()
	MWG.Add(1)
	mfx := func() {
		log.Println("lookup start")
		defer func() {
			if err := recover(); err != nil {
				log.Println("[LOOKUP error]:", err)
			}
		}()
		conf := config.GetConfig()

		//for test,only connect one
		ips := []IPPort{{
			ip:   net.ParseIP("47.97.62.19"),
			port: 8333,
		}}

		//ips := []IPPort{}
		//for _, v := range fixips {
		//	ips = append(ips, v)
		//}
		//ips = append(ips, lookupseeds(ctx, conf)...)

		ctimer := time.NewTimer(time.Millisecond * 100)
		checkAll := false
		for idx := 0; ; {
			select {
			case <-ctimer.C:
				if checkAll && OutIps.Len() >= conf.MaxOutConn {
					ctimer.Reset(time.Second * 1)
					continue
				}
				if ip, ok := getconnip(ips, idx, conf); ok {
					Addrs.Set(ip)
					IpChan <- ip
				}
				if idx++; idx >= len(ips) {
					checkAll = true
					idx = 0
					ips = Addrs.Ips()
				}
				if checkAll {
					ctimer.Reset(time.Second * 1)
				} else {
					ctimer.Reset(time.Millisecond * 100)
				}
			case <-ctx.Done():
				log.Println("LOOKUP stop", ctx.Err())
				return
			}
		}
	}
	for ctx.Err() != context.Canceled {
		time.Sleep(time.Second * 3)
		mfx()
	}
}
