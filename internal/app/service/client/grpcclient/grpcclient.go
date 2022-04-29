package grpcclient

import (
	"context"
	"fmt"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"github.com/triumphpc/GophKeeper/internal/app/service/client/authclient"
	"github.com/triumphpc/GophKeeper/internal/app/service/client/authclientinterceptor"
	"github.com/triumphpc/GophKeeper/internal/app/service/client/userdata"
	"google.golang.org/grpc"
	"time"
)

const (
	refreshDuration = 30 * time.Second // refresh token for client
)

type Client struct {
	transportOption grpc.DialOption
	cc1             *grpc.ClientConn
	cc2             *grpc.ClientConn
	ud              *userdata.UserData
}

var instance *Client

func Instance() *Client {
	if instance == nil {
		transportOption := grpc.WithInsecure()
		c := configs.ClientInstance()

		cc1, err := grpc.Dial(c.GRPCAddress, transportOption)
		if err != nil {
			c.Logger.Error(fmt.Sprint("cannot dial server: ", err))
		}

		instance = &Client{transportOption: transportOption, cc1: cc1}
	}

	return instance
}

func (client *Client) AuthClient(ctx context.Context, username string, password string) error {
	c := configs.ClientInstance()

	authClient := authclient.New(client.cc1, username, password)
	interceptor, err := authclientinterceptor.New(ctx, authClient, c.AuthMethods(), refreshDuration, false)
	if err != nil {
		return err
	}

	cc2, err := grpc.Dial(
		c.GRPCAddress,
		client.transportOption,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		return err
	}

	//Make client service
	client.ud = userdata.New(cc2)

	return nil
}

func (client *Client) RegClient(ctx context.Context, username string, password string) error {
	regClient := authclient.New(client.cc1, username, password)
	c := configs.ClientInstance()

	interceptor, err := authclientinterceptor.New(ctx, regClient, c.AuthMethods(), refreshDuration, true)
	if err != nil {
		return err
	}

	cc2, err := grpc.Dial(
		c.GRPCAddress,
		client.transportOption,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		c.Logger.Error(fmt.Sprint("cannot dial server: ", err))

		return err
	}

	client.cc2 = cc2

	return nil
}
