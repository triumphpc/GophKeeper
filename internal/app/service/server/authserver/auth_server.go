// Package authserver implement project services
package authserver

import (
	"context"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/jwt"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage"
	proto "github.com/triumphpc/GophKeeper/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer is the server for authentication
type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	userStore  storage.Storage
	jwtManager *jwt.Manager
}

// New returns a new auth server
func New(userStore storage.Storage, jwtManager *jwt.Manager) *AuthServer {
	return &AuthServer{userStore: userStore, jwtManager: jwtManager}
}

// Login is a unary RPC to login user
func (server *AuthServer) Login(_ context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	user, err := server.userStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user: %v, err %v", req.GetUsername(), err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &proto.LoginResponse{AccessToken: token}

	return res, nil
}
