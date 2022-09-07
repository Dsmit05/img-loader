package service

import (
	"context"

	"github.com/Dsmit05/img-loader/internal/models"
)

type brokerI interface {
	GetChanIMG() <-chan models.UserEvent
}

type workerPoolI interface {
	Close() error
	PushTask(ctx context.Context, task func() error) error
	ResetWorkersNum(workersNum uint) error
	ChanError() (<-chan error, error)
	GetCountWorker() uint
}

type fileLoaderI interface {
	Load(ctx context.Context, url string, path string, name string) error
	GenerateTask(ctx context.Context, url string, path string, name string) func() error
}
