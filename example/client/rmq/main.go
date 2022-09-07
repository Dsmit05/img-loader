package main

import (
	"encoding/json"
	"fmt"

	"github.com/Dsmit05/img-loader/internal/logger"
	"github.com/Dsmit05/img-loader/internal/models"
	"github.com/Dsmit05/img-loader/pkg/rmq"
)

func main() {
	if err := logger.InitLogger(false, "logs.json"); err != nil {
		panic(err)
	}

	pb := rmq.NewProducer("amqp://guest:guest@localhost:5672/")
	if err := pb.InitChannel(); err != nil {
		panic(err)
	}

	user := models.UserEvent{
		FirstName: "1",
		LastName:  "2",
		URL:       "https://img.desktopwallpapers.ru/animals/pics/wide/1920x1080/55aebb2f4aff271b807697b8cb00a5d0.jpg",
	}

	msg, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
	}

	err = pb.PublishBytes("imgE", "img-loader", msg)
	fmt.Println(err)
}
