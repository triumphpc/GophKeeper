// Server side entrypoint for service GophKeeper
package main

import (
	"context"
	"fmt"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/jwt"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage/disk"
	"github.com/triumphpc/GophKeeper/internal/app/service/server/authserver"
	"github.com/triumphpc/GophKeeper/internal/app/service/server/authserverinterceptor"
	"github.com/triumphpc/GophKeeper/internal/app/service/server/registerserver"
	"github.com/triumphpc/GophKeeper/internal/app/service/server/userdataserver"
	proto "github.com/triumphpc/GophKeeper/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	srv := rungRPCServer(stop)

	releaseResources(ctx, srv)
}

// Run gRPC server
func rungRPCServer(stop context.CancelFunc) *grpc.Server {
	c := configs.Instance()
	l, err := net.Listen("tcp", c.GRPCAddress)
	if err != nil {
		stop()
		c.Logger.Fatal(err.Error())
	}

	jwtManager := jwt.New(secretKey, tokenDuration)
	authServer := authserver.New(c.Storage, jwtManager)
	regServer := registerserver.New(c.Storage, jwtManager)
	fileStore := disk.New("tmp")

	userDataServer := userdataserver.New(c.Storage, fileStore, jwtManager)

	interceptor := authserverinterceptor.New(jwtManager, c.AccessibleRoles())
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	srv := grpc.NewServer(serverOptions...)

	// Register proto services
	proto.RegisterAuthServiceServer(srv, authServer)
	proto.RegisterRegisterServiceServer(srv, regServer)
	proto.RegisterUserDataServiceServer(srv, userDataServer)
	reflection.Register(srv)

	// get request from gRPC
	go func() {
		if err := srv.Serve(l); err != nil {
			stop()
			c.Logger.Fatal(err.Error())
		}
	}()
	c.Logger.Info("gRPC server started on " + c.GRPCAddress)

	return srv
}

// releaseResources free resources
func releaseResources(ctx context.Context, srv *grpc.Server) {
	<-ctx.Done()
	if ctx.Err() != nil {
		fmt.Printf("Error:%v\n", ctx.Err())
	}

	c := configs.Instance()

	c.Logger.Info("The service is shutting down...")

	c.Logger.Info("...closing connect to db")
	err := c.Storage.Close
	if err != nil {
		c.Logger.Info("...closing don't close")
	}
	c.Logger.Info("...db connection closed")

	// Close gRPC server
	srv.Stop()
	c.Logger.Info("...gRPC server stopped")
}
