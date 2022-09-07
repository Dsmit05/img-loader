//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Dsmit05/img-loader/pkg/api"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestIMGLoader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		go FileServer.SetUp(t)
		defer FileServer.TearDown(t)

		resp, err := UserImageClient.PutImage(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		msg := api.PutImageRequest{
			Url:       "http://localhost:9000/static/img.jpg",
			FirstName: "D",
			LastName:  "Smit",
		}

		if err := resp.Send(&msg); err != nil {
			t.Fatal(err)
		}

		_, err = resp.CloseAndRecv()
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second * 3)

		fs, err := os.Stat("../Users/D_Smit/img.jpg")
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("file does not exist: %v", err)
		}

		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, "img.jpg", fs.Name())
	})

}
