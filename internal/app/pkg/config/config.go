// Package configs implement functions for project configs
package configs

import (
	"context"
	"encoding/json"
	_ "github.com/caarlos0/env/v6"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage/pg"
	"github.com/triumphpc/GophKeeper/pkg/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
)

// Config project
type Config struct {
	GRPCAddress string

	Storage storage.Storage
	Logger  *zap.Logger
}

// JSONConfig for json config
type JSONConfig struct {
	GRPCAddress string `json:"grpc_address"`
	DatabaseDsn string `json:"database_dsn"`
}

// ConfigPath Config project path
const ConfigPath = "/configs/env.json"

var instance *Config

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

	var config JSONConfig

	// jsonFile's content into 'config' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		lgr.Fatal("Can't init project configuration")
	}

	instance = new(Config)
	instance.GRPCAddress = config.GRPCAddress
	instance.Logger = lgr

	stg, err := pg.New(context.Background(), lgr, config.DatabaseDsn)
	if err != nil {
		log.Fatal(err)
	}

	instance.Storage = stg
}

// Instance return singleton
func Instance() *Config {
	return instance
}
