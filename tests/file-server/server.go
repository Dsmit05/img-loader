//go:build integration
// +build integration

package file_server

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Dsmit05/img-loader/internal/logger"
)

type FileServer struct {
	s *http.Server
	sync.Mutex
}

func NewFileServer() *FileServer {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./file-server/file"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	s := &http.Server{
		Addr:    ":9000",
		Handler: mux,
	}

	return &FileServer{s: s}
}

func (f *FileServer) SetUp(t *testing.T) {
	t.Helper()
	f.Lock()
	if err := f.s.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			t.Log("FileServer Server closed under request")
		} else {
			t.Errorf("FileServer closed unexpect with err: %v", err)
		}
	}
}

func (f *FileServer) TearDown(t *testing.T) {
	defer f.Unlock()
	stop := make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go func() {
		f.s.SetKeepAlivesEnabled(false)
		if err := f.s.Shutdown(ctx); err != nil {
			logger.Error("FileServer Shutdown failed", err)
		}
		stop <- true
	}()

	select {
	case <-ctx.Done():
		t.Errorf("FileServer context timeout", ctx.Err())
	case <-stop:
		t.Log("FileServer stop")
	}
}
