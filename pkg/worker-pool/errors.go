package worker_pool

import "github.com/pkg/errors"

var (
	ErrWorkerZeroCount = errors.New("count of workers must be greater than 0")
	ErrChanErrIsNil    = errors.New("chanel error is nil")
)
