//go:build integration
// +build integration

package tests

import (
	"time"

	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/Dsmit05/img-loader/pkg/api"
	"github.com/Dsmit05/img-loader/tests/config"
	file_server "github.com/Dsmit05/img-loader/tests/file-server"
	"google.golang.org/grpc"
)

var (
	UserImageClient api.UserImageClient
	FileServer      *file_server.FileServer
)

func init() {
	_ = logger.InitLogger(true, "")
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Fatal("config.FromEnv()", err)
	}

	conn, err := grpc.Dial(cfg.Host, grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		logger.Fatal("grpc.Dial()", err)
	}

	FileServer = file_server.NewFileServer()
	UserImageClient = api.NewUserImageClient(conn)
}
