package main

import (
	"context"
	"fmt"
	"github.com/teneta-io/dcp/internal/container"
	"github.com/teneta-io/dcp/internal/http"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
	}()

	c := container.Build(ctx)
	s := c.Get("Server").(*http.Server)
	l := c.Get("Logger").(*zap.Logger)

	go s.Run()

	l.Info(fmt.Sprintf("Got %s signal. Shutting down...", <-waitSigTerm()))
}

func waitSigTerm() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT)
	return signalChan
}
