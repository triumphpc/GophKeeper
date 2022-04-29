package userdata

import (
	"bufio"
	"context"
	"fmt"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	proto "github.com/triumphpc/GophKeeper/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"os"
	"path/filepath"
	"time"
)

// UserData is a client to call service RPCs
type UserData struct {
	service proto.UserDataServiceClient
}

// New returns a new service client
func New(cc *grpc.ClientConn) *UserData {
	service := proto.NewUserDataServiceClient(cc)

	return &UserData{service}
}

// SaveText calls create user text data by RPC
func (client *UserData) SaveText(text *proto.Text) {
	req := &proto.SaveTextRequest{
		Text: text,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.service.SaveText(ctx, req)
	c := configs.Instance()

	if err != nil {
		c.Logger.Fatal("cannot create text: " + err.Error())
	}

	c.Logger.Info("created text with id: " + res.GetId())
}

// SaveCard calls create user card data by RPC
func (client *UserData) SaveCard(text *proto.Card) {
	req := &proto.SaveCardRequest{
		Card: text,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.service.SaveCard(ctx, req)
	c := configs.Instance()

	if err != nil {
		c.Logger.Fatal("cannot create card: " + err.Error())
	}

	c.Logger.Info("created card with id: " + res.GetId())
}

// UploadFile calls upload file RPC
func (client *UserData) UploadFile(filePath string, meta string) {
	file, err := os.Open(filePath)
	c := configs.Instance()

	if err != nil {
		c.Logger.Fatal("cannot open file: " + err.Error())
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.UploadFile(ctx)
	if err != nil {
		c.Logger.Fatal("cannot open file: " + err.Error())
	}

	req := &proto.UploadFileRequest{
		Data: &proto.UploadFileRequest_Info{
			Info: &proto.FileInfo{
				FileType: filepath.Ext(filePath),
				Meta:     meta,
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		c.Logger.Fatal("cannot send image info to server: ", zap.Error(err), zap.Error(stream.RecvMsg(nil)))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.Logger.Fatal("cannot read chunk to buffer: ", zap.Error(err))
		}

		req := &proto.UploadFileRequest{
			Data: &proto.UploadFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			c.Logger.Fatal("cannot send chunk to server: ", zap.Error(err), zap.Error(stream.RecvMsg(nil)))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		c.Logger.Fatal("cannot receive response: ", zap.Error(err))
	}

	c.Logger.Info(fmt.Sprintf("file uploaded with id: %s, size: %d ", res.GetId(), res.GetSize()))
}
