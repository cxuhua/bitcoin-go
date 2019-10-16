package main

import (
	"bitcoin/core"
	"bitcoin/db"
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
	err := db.UseSession(ctx, func(db db.DbImp) error {
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
