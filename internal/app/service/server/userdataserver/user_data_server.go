package userdataserver

import (
	"bytes"
	"context"
	"fmt"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/jwt"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage"
	"github.com/triumphpc/GophKeeper/internal/app/service/userdata"
	proto "github.com/triumphpc/GophKeeper/pkg/api"
	"github.com/triumphpc/GophKeeper/pkg/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"strconv"
)

// maxFileSize set max file upload size
const maxFileSize = 1 << 20

// UserDataServer is the server that provides client services
type UserDataServer struct {
	proto.UnimplementedUserDataServiceServer
	userStore  storage.Storage
	fileStore  storage.FileStorage
	jwtManager *jwt.Manager
}

// New returns a new LaptopServer
func New(userStore storage.Storage, fileStore storage.FileStorage, jwtManager *jwt.Manager) *UserDataServer {
	return &UserDataServer{
		userStore:  userStore,
		fileStore:  fileStore,
		jwtManager: jwtManager,
	}
}

// SaveText implement save text type data
func (server *UserDataServer) SaveText(ctx context.Context, req *proto.SaveTextRequest) (*proto.SaveTextResponse, error) {
	textData := userdata.NewDataText(
		req.GetText().GetName(),
		req.GetText().GetData(),
		req.GetText().GetMeta(),
	)

	err := server.userStore.SaveText(ctx, textData, server.jwtManager.Claims().Id)
	if err != nil {
		return nil, err
	}

	res := &proto.SaveTextResponse{Id: strconv.Itoa(textData.Id)}

	return res, nil
}

// SaveCard implement save card type data
func (server *UserDataServer) SaveCard(ctx context.Context, req *proto.SaveCardRequest) (*proto.SaveCardResponse, error) {
	number, err := crypto.Encode(req.GetCard().GetNumber())
	if err != nil {
		return nil, err
	}

	cardData := userdata.NewDataCard(
		number,
		req.GetCard().GetMeta(),
	)

	err = server.userStore.SaveCard(ctx, cardData, server.jwtManager.Claims().Id)
	if err != nil {
		return nil, err
	}

	res := &proto.SaveCardResponse{Id: strconv.Itoa(cardData.Id)}

	return res, nil
}

// UploadFile is a client-streaming RPC to upload a laptop image
func (server *UserDataServer) UploadFile(stream proto.UserDataService_UploadFileServer) error {
	req, err := stream.Recv()
	c := configs.Instance()

	if err != nil {
		c.Logger.Error(status.Errorf(codes.Unknown, "cannot receive image info").Error())
	}

	fileType := req.GetInfo().GetFileType()
	c.Logger.Info(fmt.Sprintf("receive an upload-image request with image type %s", fileType))

	fileData := bytes.Buffer{}
	fileSize := 0

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		c.Logger.Info("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			c.Logger.Info("no more data")
			break
		}
		if err != nil {
			c.Logger.Error(status.Errorf(codes.Unknown, "cannot receive chunk data").Error())
			return err
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		c.Logger.Info(fmt.Sprintf("received a chunk with size: %d", size))

		fileSize += size
		if fileSize > maxFileSize {
			c.Logger.Error(status.Errorf(codes.Unknown, "image is too large: %d > %d", fileSize, maxFileSize).Error())
			return err
		}

		_, err = fileData.Write(chunk)
		if err != nil {
			c.Logger.Error(status.Errorf(codes.Internal, "cannot write chunk data: %v", err).Error())
			return err
		}
	}

	fileInfo, err := server.fileStore.Save(stream.Context(), fileType, fileData)
	if err != nil {
		c.Logger.Error(status.Errorf(codes.Internal, "cannot save image to the store: %v", err).Error())
		return err
	}

	res := &proto.UploadFileResponse{
		Id:   fileInfo.ID.String(),
		Size: uint32(fileSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		c.Logger.Error(status.Errorf(codes.Unknown, "cannot send response: %v", err).Error())
		return err
	}

	c.Logger.Info(fmt.Sprintf("saved image with id: %s, size: %d", fileInfo.ID.String(), fileSize))

	// Create relation
	fileInfo.Meta = req.GetInfo().GetMeta()
	err = server.userStore.SaveFile(stream.Context(), *fileInfo, server.jwtManager.Claims().Id)
	if err != nil {
		return err
	}

	return nil
}

// contextError parse context
func contextError(ctx context.Context) error {
	c := configs.Instance()

	switch ctx.Err() {
	case context.Canceled:
		c.Logger.Error(status.Errorf(codes.Canceled, "request is canceled").Error())
	case context.DeadlineExceeded:
		c.Logger.Error(status.Errorf(codes.DeadlineExceeded, "deadline is exceeded").Error())
	}

	return nil
}
