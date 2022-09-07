package main

import "github.com/Dsmit05/img-loader/internal/logger"

func main() {
	produser, err := NewProducer("localhost:19091", "img")
	if err != nil {
		logger.Fatal("broker.NewProducer()", err)
	}
	produser.Start()

	if err := produser.Send(UserEvent{
		FirstName: "first",
		LastName:  "name",
		URL:       "https://img.desktopwallpapers.ru/animals/pics/wide/1920x1080/55aebb2f4aff271b807697b8cb00a5d0.jpg",
	}); err != nil {
		logger.Fatal("produser.Send()", err)
	}
	produser.Stop()
}
