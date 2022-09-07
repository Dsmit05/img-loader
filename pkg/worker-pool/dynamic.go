package worker_pool

import (
	"context"
	"sync"

	"go.uber.org/atomic"
)

type DynamicWorkersPool struct {
	tasks chan func() error
	err   chan error

	countWorker *atomic.Uint64
	mx          sync.Mutex
	wg          *sync.WaitGroup
	quit        chan struct{}
}

func NewDynamicWorkersPool(countWorker, taskPoolSize uint, checkErr bool) (*DynamicWorkersPool, error) {
	if countWorker == 0 {
		return nil, ErrWorkerZeroCount
	}

	var chErr chan error
	if checkErr {
		chErr = make(chan error, taskPoolSize)
	}

	wp := &DynamicWorkersPool{
		countWorker: atomic.NewUint64(0),
		tasks:       make(chan func() error, taskPoolSize),
		wg:          new(sync.WaitGroup),
		quit:        make(chan struct{}),
		err:         chErr,
	}

	_ = wp.ResetWorkersNum(countWorker)
	return wp, nil
}

func (dwp *DynamicWorkersPool) Close() error {
	close(dwp.tasks)
	dwp.wg.Wait()
	close(dwp.quit)
	dwp.countWorker.Store(0)
	if dwp.err != nil {
		close(dwp.err)
	}

	return nil
}

func (dwp *DynamicWorkersPool) PushTask(ctx context.Context, task func() error) error {
	select {
	case dwp.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (dwp *DynamicWorkersPool) ResetWorkersNum(workersNum uint) error {
	if workersNum == 0 {
		return ErrWorkerZeroCount
	}

	dwp.mx.Lock()
	defer dwp.mx.Unlock()

	curr := uint(dwp.countWorker.Load())
	if workersNum > curr {
		dwp.increaseWorkers(workersNum - curr)
	} else if workersNum < curr {
		dwp.decreaseWorkers(curr - workersNum)
	}

	return nil
}

func (dwp *DynamicWorkersPool) ChanError() (<-chan error, error) {
	if dwp.err == nil {
		return nil, ErrChanErrIsNil
	}

	return dwp.err, nil
}

func (dwp *DynamicWorkersPool) GetCountWorker() uint {
	return uint(dwp.countWorker.Load())
}

func (dwp *DynamicWorkersPool) worker() {
	defer dwp.wg.Done()
	for {
		select {
		case <-dwp.quit:
			return
		case task, ok := <-dwp.tasks:
			if ok {
				if err := task(); err != nil && dwp.err != nil {
					dwp.err <- err
				}
			} else {
				return
			}
		}
	}
}

func (dwp *DynamicWorkersPool) decreaseWorkers(delta uint) {
	curr := uint(dwp.countWorker.Load())
	if delta > curr {
		delta = curr
	}

	dwp.countWorker.Sub(uint64(delta))
	for i := uint(0); i < delta; i++ {
		dwp.quit <- struct{}{}
	}
}

func (dwp *DynamicWorkersPool) increaseWorkers(delta uint) {
	dwp.countWorker.Add(uint64(delta))
	dwp.wg.Add(int(delta))

	for i := uint(0); i < delta; i++ {
		go dwp.worker()
	}
}
