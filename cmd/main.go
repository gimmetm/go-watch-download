package main

import (
	"context"
	"github.com/gimmetm/go-run-download/pkg/fileworker"
	log "github.com/gimmetm/go-run-download/pkg/logging"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	log.Logger.Infoln("Starting Download Watcher.")

	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	fw := fileworker.New()
	fw.Start(ctx, wg)

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan // Blocks here until interrupted

	log.Logger.Debugln("Shutdown signal received : Wait File WatcherEnd.")
	cancelFunc()
	wg.Wait()

	log.Logger.Infoln("Stop Download Watcher.")
	os.Exit(0)

}
