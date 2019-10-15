package main

import (
	"bitcoin/core"
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
	go core.StartLookUp(ctx)
	//startup block sync
	go core.StartDispatch(ctx)
	//start worker
	go core.StartWorker(ctx, 4)
	//wait quit
	signal.Notify(csig, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	sig := <-csig
	log.Println("recv sig :", sig, ",system wait exit")
	cancel()
	core.MWG.Wait()
	log.Println("recv sig :", sig, ",system exited")
}
