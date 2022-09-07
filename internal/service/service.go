package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dsmit05/img-loader/internal/models"

	"github.com/Dsmit05/img-loader/internal/logger"
)

const (
	workerMax uint = 15
	workerMin uint = 3
)

type SaveLoader struct {
	pool       workerPoolI
	fileLoader fileLoaderI
	brokers    []brokerI
}

func NewSaveLoader(pool workerPoolI, fileLoader fileLoaderI, broker ...brokerI) (*SaveLoader, error) {
	chPoolErr, err := pool.ChanError()
	if err != nil {
		return nil, err
	}

	go func() {
		for err := range chPoolErr {
			logger.Error("error from chPoolErr", err)
		}
	}()

	return &SaveLoader{brokers: broker, pool: pool, fileLoader: fileLoader}, nil
}

func (s *SaveLoader) Start() {
	for _, brokers := range s.brokers {
		go s.process(brokers.GetChanIMG())
	}

}

func (s *SaveLoader) process(ch <-chan models.UserEvent) {
	var timeWorker = time.Now()

	for imgInfo := range ch {
		timeStart := time.Now()

		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

		path := fmt.Sprintf("%v_%v", imgInfo.FirstName, imgInfo.LastName)
		task := s.fileLoader.GenerateTask(ctx, imgInfo.URL, path, "")

		err := s.pool.PushTask(ctx, task)
		if err != nil {
			logger.Error("pool.PushTask()", err)
		}

		timeSince := time.Since(timeStart)
		if timeSince > time.Second && s.pool.GetCountWorker() < workerMax {
			s.pool.ResetWorkersNum(s.pool.GetCountWorker() + 1)
			timeWorker = time.Now()
		}

		// if the load has dropped, we gradually reduce the number of workers
		if time.Since(timeWorker) > time.Second*30 && s.pool.GetCountWorker() > workerMin {
			// Todo: add increment decrement functions for the worker pool
			s.pool.ResetWorkersNum(s.pool.GetCountWorker() - 1)
		}
	}

}

func (s *SaveLoader) Stop() {
	s.pool.Close()
}
