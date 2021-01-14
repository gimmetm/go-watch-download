package main

import (
	"context"
	"github.com/gimmetm/go-run-download/pkg/fileworker"
	log "github.com/gimmetm/go-run-download/pkg/logging"
	"sync"
)

func main(){

	log.Logger.Infoln("Starting Download Watcher.")


	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)

	fw := fileworker.New()
	fw.Start(ctx, wg)

	wg.Wait()
	ctx.Done()
	log.Logger.Infoln("Stop Download Watcher.")
}
