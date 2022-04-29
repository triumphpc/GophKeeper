package authclientinterceptor

import (
	"context"
	"fmt"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"github.com/triumphpc/GophKeeper/internal/app/service/client/authclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

// AuthInterceptor is a client interceptor for authentication
type AuthInterceptor struct {
	authClient  *authclient.AuthClient
	authMethods map[string]bool
	accessToken string
	isReg       bool
}

// New returns a new auth interceptor
func New(
	ctx context.Context,
	authClient *authclient.AuthClient,
	authMethods map[string]bool,
	refreshDuration time.Duration,
	isRegistration bool,
) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient:  authClient,
		authMethods: authMethods,
		isReg:       isRegistration,
	}

	err := interceptor.scheduleRefreshToken(ctx, refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

// scheduleRefreshToken update token by schedule
func (interceptor *AuthInterceptor) scheduleRefreshToken(ctx context.Context, refreshDuration time.Duration) error {
	err := interceptor.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			select {
			case <-time.After(wait):
				err := interceptor.refreshToken()
				if err != nil {
					wait = time.Second
				} else {
					wait = refreshDuration
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// refreshToken for user
func (interceptor *AuthInterceptor) refreshToken() error {
	var accessToken string
	var err error

	if interceptor.isReg {
		accessToken, err = interceptor.authClient.Registration()
	} else {
		accessToken, err = interceptor.authClient.Login()
	}
	if err != nil {
		return err
	}

	interceptor.accessToken = accessToken
	configs.Instance().Logger.Info(fmt.Sprintf("token refreshed: %v", accessToken))

	return nil
}

// attachToken to request context
func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
}

// Unary returns a client interceptor to authenticate unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		configs.Instance().Logger.Info(fmt.Sprintf("--> unary interceptor: %s", method))

		if interceptor.authMethods[method] {
			return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// Stream returns a client interceptor to authenticate stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		configs.Instance().Logger.Info(fmt.Sprintf("--> stream interceptor: %s", method))

		if interceptor.authMethods[method] {
			return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}
