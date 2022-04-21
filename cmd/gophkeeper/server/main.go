package main

import (
	"context"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Init context
	_, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	srv := grpc.NewServer()

	rungRPCServer(srv, stop)

}

// Run gRPC server
func rungRPCServer(s *grpc.Server, stop context.CancelFunc) {
	c := configs.Instance()

	l, err := net.Listen("tcp", c.GRPCAddress)
	if err != nil {
		c.Logger.Fatal(err.Error())
		stop()

	}
	c.Logger.Info("gRPC server started on " + c.GRPCAddress)

	// get request from gRPC
	go func() {
		if err := s.Serve(l); err != nil {
			stop()
			c.Logger.Fatal(err.Error())
		}
	}()
}
