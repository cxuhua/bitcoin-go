package main

import (
	"bitcoin/core"
	"bitcoin/store"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	csig := make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())
	//init global value
	err := store.UseSession(ctx, func(db store.DbImp) error {
		return core.G.Init(db)
	})
	if err != nil {
		log.Println("init global value error", err)
		os.Exit(-1)
		return
	}
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
