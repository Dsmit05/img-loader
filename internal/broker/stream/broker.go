package stream

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Dsmit05/img-loader/internal/config"
	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/Dsmit05/img-loader/internal/models"
	pb "github.com/Dsmit05/img-loader/pkg/api"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ImageListener struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	cfg        *config.Config
	userIS     *LoaderServer
}

func NewImageListener(cfg *config.Config) *ImageListener {
	loadServer := NewLoaderServer()

	return &ImageListener{cfg: cfg, userIS: loadServer}
}

func (i *ImageListener) GetChanIMG() <-chan models.UserEvent {
	return i.userIS.GetChanIMG()
}

func (i *ImageListener) StartGRPC() {
	listener, err := net.Listen("tcp", i.cfg.GetApiGRPCServerAddr())
	if err != nil {
		logger.Error("net.Listen() error", err)
		return
	}

	newGrpcServer := grpc.NewServer(grpc.ConnectionTimeout(i.cfg.GetApiGRPCServerTimeout()))
	i.grpcServer = newGrpcServer
	pb.RegisterUserImageServer(i.grpcServer, i.userIS)

	if err = i.grpcServer.Serve(listener); err != nil {
		logger.Error("grpcServer.Serve() error", err)
	}
}

func (i *ImageListener) StartREST() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := pb.RegisterUserImageHandlerFromEndpoint(ctx, mux, i.cfg.GetApiGRPCServerAddr(), opts); err != nil {
		logger.Error("RegisterUserImageHandlerFromEndpoint() error", err)
		return
	}

	newHttpServer := &http.Server{
		Addr:    i.cfg.GetApiHTTPServerAddr(),
		Handler: mux,
	}

	i.httpServer = newHttpServer

	if err := i.httpServer.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			logger.Info("ApiServer", "Server closed under request")
		} else {
			logger.Error("http.ListenAndServe() error", err)
		}
	}

}

func (i *ImageListener) Stop() {
	stop := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go func() {
		i.httpServer.SetKeepAlivesEnabled(false)
		if err := i.httpServer.Shutdown(ctx); err != nil {
			logger.Error("Server Shutdown failed", err)
			return
		}

		i.grpcServer.GracefulStop()

		stop <- true
	}()

	select {
	case <-ctx.Done():
		i.grpcServer.Stop()
		logger.Error("Server context timeout", ctx.Err())
	case <-stop:
		logger.Info("Server", "Server closed under request")
	}

	i.userIS.CloseChanIMG()
}
