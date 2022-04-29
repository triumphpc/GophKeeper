// Package configs implement functions for project configs
package configs

import (
	"encoding/json"
	_ "github.com/caarlos0/env/v6"
	"github.com/triumphpc/GophKeeper/pkg/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
)

// ClientConfig project
type ClientConfig struct {
	GRPCAddress string
	Logger      *zap.Logger
}

// JSONClientConfig for json config
type JSONClientConfig struct {
	GRPCAddress string `json:"grpc_address"`
}

// ClientConfigPath Config project path
const ClientConfigPath = "/configs/client_env.json"

var clientInstance *ClientConfig

func init() {
	lgr, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	// Init from json evn config
	pwd, _ := os.Getwd()
	path := pwd + ConfigPath
	byteValue, err := ioutil.ReadFile(path)
	if err != nil {
		lgr.Fatal("Can't init project configuration")
	}

	var config JSONClientConfig

	// jsonFile's content into 'config' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		lgr.Fatal("Can't init project configuration")
	}

	clientInstance = new(ClientConfig)
	clientInstance.GRPCAddress = config.GRPCAddress
	clientInstance.Logger = lgr

}

// ClientInstance return singleton
func ClientInstance() *ClientConfig {
	return clientInstance
}

// AuthMethods methods for client roles
func (c *ClientConfig) AuthMethods() map[string]bool {
	return map[string]bool{
		userDataServicePath + "SaveText": true,
	}
}
