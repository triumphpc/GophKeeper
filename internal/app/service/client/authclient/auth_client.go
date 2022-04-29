// Package authclient implement gRPC client
package authclient

import (
	"context"
	proto "github.com/triumphpc/GophKeeper/pkg/api"
	"google.golang.org/grpc"
	"time"
)

// AuthClient is a client to call authentication RPC
type AuthClient struct {
	service    proto.AuthServiceClient
	serviceReg proto.RegisterServiceClient
	username   string
	password   string
}

// New returns a new auth client
func New(cc *grpc.ClientConn, username string, password string) *AuthClient {
	service := proto.NewAuthServiceClient(cc)
	serviceReg := proto.NewRegisterServiceClient(cc)
	return &AuthClient{service, serviceReg, username, password}
}

// Login user and returns the access token
func (client *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.LoginRequest{
		Username: client.username,
		Password: client.password,
	}

	res, err := client.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}

// Registration create user and returns the access token
func (client *AuthClient) Registration() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.RegisterRequest{
		Username: client.username,
		Password: client.password,
		Role:     "user",
	}

	res, err := client.serviceReg.Register(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
