package client

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/tonx22/playlist/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

type PlaylistService interface {
	Play(id int32) error
	Pause() error
	Next() error
	Prev() error
	AddSong(song *Song) (*Song, error)
	Delete(id int32) error
	GetPlaylist() error
}

type playlistService struct {
	GRPCClient pb.PlaylistSvcClient
}

func NewGRPCClient() (*playlistService, error) {
	defaultHost, ok := os.LookupEnv("GRPC_HOST")
	if !ok {
		defaultHost = "localhost"
	}
	defaultPort, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		defaultPort = "50051"
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	serverAddr := fmt.Sprintf("%s:%s", defaultHost, defaultPort)
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("fail to dial: %s", err))
	}
	svc := playlistService{GRPCClient: pb.NewPlaylistSvcClient(conn)}
	return &svc, nil
}

func (svc playlistService) Play(id int32) error {
	_, err := svc.GRPCClient.Play(context.TODO(), &pb.PlayRequest{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (svc playlistService) Pause() error {
	_, err := svc.GRPCClient.Pause(context.TODO(), &pb.EmptyParams{})
	if err != nil {
		return err
	}
	return nil
}

func (svc playlistService) Next() error {
	_, err := svc.GRPCClient.Next(context.TODO(), &pb.EmptyParams{})
	if err != nil {
		return err
	}
	return nil
}

func (svc playlistService) Prev() error {
	_, err := svc.GRPCClient.Prev(context.TODO(), &pb.EmptyParams{})
	if err != nil {
		return err
	}
	return nil
}

func (svc playlistService) AddSong(song *Song) (*Song, error) {
	resp, err := svc.GRPCClient.AddSong(context.TODO(), &pb.Song{Description: song.Description, Duration: song.Duration})
	if err != nil {
		return nil, err
	}
	song.Id = resp.Id
	return song, nil
}

func (svc playlistService) Delete(id int32) error {
	_, err := svc.GRPCClient.Delete(context.TODO(), &pb.PlayRequest{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (svc playlistService) GetPlaylist() error {
	_, err := svc.GRPCClient.GetPlaylist(context.TODO(), &pb.EmptyParams{})
	if err != nil {
		return err
	}
	return nil
}

type Song struct {
	Id          int32  `json:"id"`
	Description string `json:"description"`
	Duration    int32  `json:"duration"`
	Prev        int32  `json:"prev"`
	Next        int32  `json:"next"`
}
