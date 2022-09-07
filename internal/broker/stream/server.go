package stream

import (
	"io"

	"github.com/Dsmit05/img-loader/internal/models"
	pb "github.com/Dsmit05/img-loader/pkg/api"
)

type LoaderServer struct {
	pb.UnimplementedUserImageServer
	chImg chan models.UserEvent
}

func NewLoaderServer() *LoaderServer {
	chIMG := make(chan models.UserEvent)
	return &LoaderServer{chImg: chIMG}
}

func (l *LoaderServer) PutImage(stream pb.UserImage_PutImageServer) error {
	for {
		imgInfo, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.PutImageResponse{})
		}

		if err != nil {
			return err
		}

		l.chImg <- models.UserEvent{
			FirstName: imgInfo.GetFirstName(),
			LastName:  imgInfo.GetLastName(),
			URL:       imgInfo.GetUrl(),
		}

	}
}

func (l *LoaderServer) GetChanIMG() <-chan models.UserEvent {
	return l.chImg
}

func (l *LoaderServer) CloseChanIMG() {
	close(l.chImg)
}
