package transport

import (
	"context"
	"fmt"
	pb "github.com/tonx22/playlist/pb"
	Models "github.com/tonx22/playlist/pkg/models"
	"github.com/tonx22/playlist/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"time"
)

type server struct {
	pb.UnimplementedPlaylistSvcServer
	service service.PlaylistService
}

func StartNewGRPCServer(svc service.PlaylistService, grpcPort int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPlaylistSvcServer(grpcServer, &server{service: svc})

	ch := make(chan error)
	go func() {
		ch <- grpcServer.Serve(lis)
	}()

	var e error
	select {
	case e = <-ch:
		return e
	case <-time.After(time.Second * 1):
	}
	log.Printf("GRPC server listening at %v", lis.Addr())
	return nil
}

func (s *server) Play(ctx context.Context, in *pb.PlayRequest) (*pb.EmptyParams, error) {
	err := s.service.Play(int(in.Id))
	if err != nil {
		return &pb.EmptyParams{}, encodeGRPCError(err)
	}
	return &pb.EmptyParams{}, nil
}

func (s *server) Pause(ctx context.Context, in *pb.EmptyParams) (*pb.EmptyParams, error) {
	err := s.service.Pause()
	if err != nil {
		return &pb.EmptyParams{}, encodeGRPCError(err)
	}
	return &pb.EmptyParams{}, nil
}

func (s *server) Next(ctx context.Context, in *pb.EmptyParams) (*pb.EmptyParams, error) {
	err := s.service.Next()
	if err != nil {
		return &pb.EmptyParams{}, encodeGRPCError(err)
	}
	return &pb.EmptyParams{}, nil
}

func (s *server) Prev(ctx context.Context, in *pb.EmptyParams) (*pb.EmptyParams, error) {
	err := s.service.Prev()
	if err != nil {
		return &pb.EmptyParams{}, encodeGRPCError(err)
	}
	return &pb.EmptyParams{}, nil
}

func (s *server) AddSong(ctx context.Context, in *pb.Song) (*pb.Song, error) {
	res, err := s.service.AddSong(&Models.Song{Description: in.Description, Duration: int(in.Duration)})
	if err != nil {
		return nil, encodeGRPCError(err)
	}
	in.Id = int32(res.Id)
	return in, nil
}

func (s *server) Delete(ctx context.Context, in *pb.PlayRequest) (*pb.EmptyParams, error) {
	err := s.service.Delete(int(in.Id))
	if err != nil {
		return &pb.EmptyParams{}, encodeGRPCError(err)
	}
	return &pb.EmptyParams{}, nil
}

func (s *server) GetPlaylist(ctx context.Context, in *pb.EmptyParams) (*pb.PlaylistReply, error) {
	songs, err := s.service.GetPlaylist()
	if err != nil {
		return nil, encodeGRPCError(err)
	}
	res := make([]*pb.Song, 0)
	for _, s := range *songs {
		res = append(res, &pb.Song{Id: int32(s.Id), Description: s.Description, Duration: int32(s.Duration), Prev: int32(s.Prev), Next: int32(s.Next)})
	}
	return &pb.PlaylistReply{Songs: res}, nil
}

func encodeGRPCError(err error) error {
	e, ok := err.(Models.ResponseError)
	if !ok {
		return status.Error(codes.Internal, err.Error())
	}
	var code codes.Code
	switch e.Status {
	case http.StatusNotFound:
		code = codes.NotFound
	default:
		code = codes.Internal
	}
	return status.Error(code, e.Error())
}
