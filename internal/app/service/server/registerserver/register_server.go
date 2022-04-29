// Package registerserver implement registration gRPC server
package registerserver

import (
	"context"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/jwt"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage"
	userservice "github.com/triumphpc/GophKeeper/internal/app/service/user"
	proto "github.com/triumphpc/GophKeeper/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterServer is the server for registration
type RegisterServer struct {
	proto.UnimplementedRegisterServiceServer
	userStore  storage.Storage
	jwtManager *jwt.Manager
}

// New returns a new auth server
func New(userStore storage.Storage, jwtManager *jwt.Manager) *RegisterServer {
	return &RegisterServer{userStore: userStore, jwtManager: jwtManager}
}

// Register registration user logic
func (server *RegisterServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	_, err := server.userStore.Find(req.GetUsername())
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "login already exist: %v", err)
	}

	// Create new user
	user, err := userservice.New(req.GetUsername(), req.GetPassword(), req.GetRole())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create hash for new user")
	}

	if err = server.userStore.CreateUser(user); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create user")
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &proto.RegisterResponse{AccessToken: token}

	return res, nil

}
