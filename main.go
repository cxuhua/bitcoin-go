package main

import (
	"bitcoin/net"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	csig := make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())
	//startup lookup
	go net.StartLookUp(ctx)
	//startup block sync
	go net.StartRsync(ctx)
	//wait quit
	signal.Notify(csig, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	sig := <-csig
	log.Println("recv sig :", sig, ",system wait exit")
	cancel()
	net.MWG.Wait()
	log.Println("recv sig :", sig, ",system exited")
}
