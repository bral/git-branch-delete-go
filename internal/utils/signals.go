package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithSignals returns a context that is canceled when the program receives
// an interrupt signal (SIGINT, SIGTERM)
func WithSignals(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigCh)
	}()

	return ctx
}

// HandleSignals sets up signal handling and returns a cleanup function
func HandleSignals(cleanup func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		if cleanup != nil {
			cleanup()
		}
		os.Exit(1)
	}()
}
