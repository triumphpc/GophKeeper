// Package implement client for service
package main

import (
	"context"
	"fmt"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/tui"
	"google.golang.org/grpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	refreshDuration = 30 * time.Second // refresh token for client
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	tui.InitialModel(ctx) // Run TUI

	//transportOption := grpc.WithInsecure()
	//c := configs.ClientInstance()
	//
	//cc1, err := grpc.Dial(c.GRPCAddress, transportOption)
	//if err != nil {
	//	c.Logger.Error(fmt.Sprint("cannot dial server: ", err))
	//}

	//username, password := "login", "password"
	//
	//authClient := authclient.New(cc1, username, password)
	//interceptor, err := authclientinterceptor.New(ctx, authClient, c.AuthMethods(), refreshDuration)
	//if err != nil {
	//	c.Logger.Error(fmt.Sprint("cannot create auth interceptor: ", err))
	//}
	////
	//cc2, err := grpc.Dial(
	//	c.GRPCAddress,
	//	transportOption,
	//	grpc.WithUnaryInterceptor(interceptor.Unary()),
	//	grpc.WithStreamInterceptor(interceptor.Stream()),
	//)
	//if err != nil {
	//	c.Logger.Error(fmt.Sprint("cannot dial server: ", err))
	//}
	//
	////Make client service
	//userData := userdata.New(cc2)
	//
	//userData.SaveText(&proto.Text{Name: "Test"})
	//userData.SaveCard(&proto.Card{Number: "1234123412341234"})
	//userData.UploadFile("tmp/laptop.jpg", "Meta info")
	//
	//releaseResources(ctx, cc1, cc2)

}

// userCredential get user credential
//func userCredential() (username string, password string) {
//	reader := bufio.NewReader(os.Stdin)
//
//	fmt.Print("Enter username: ")
//	username, _ = reader.ReadString('\n')
//	username = strings.Trim(username, "\n")
//
//	fmt.Print("Enter password: ")
//	password, _ = reader.ReadString('\n')
//	password = strings.Trim(password, "\n")
//
//	return
//}

// releaseResources free resources
func releaseResources(ctx context.Context, сс1 *grpc.ClientConn, cc2 *grpc.ClientConn) {
	<-ctx.Done()

	if ctx.Err() != nil {
		fmt.Printf("Error: %v\n", ctx.Err())
	}

	c := configs.Instance()

	c.Logger.Info("The client is shutting down...")
	// Close gRPC clients
	сс1.Close()
	cc2.Close()

	c.Logger.Info("...gRPC client stopped")
}
