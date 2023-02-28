package client

import (
	"fmt"
	"testing"
	"time"
)

func TestGRPCClient(t *testing.T) {

	client, err := NewGRPCClient()
	if err != nil {
		t.Fatalf("failed to create grpc client: %v", err)
	}

	t.Run("(1) начало воспроизведения", func(t *testing.T) {
		err = client.Play(0)
		if err != nil {
			t.Fatalf("failed to song playing: %v", err)
		}
	})
	time.Sleep(time.Second * 4)

	t.Run("(2) приостановка воспроизведения", func(t *testing.T) {
		err = client.Pause()
		if err != nil {
			t.Fatalf("failed to pause of playing the song: %v", err)
		}
	})
	time.Sleep(time.Second * 4)

	t.Run("(3) возобновление воспроизведения", func(t *testing.T) {
		err = client.Play(0)
		if err != nil {
			t.Fatalf("failed to continue of playing the song: %v", err)
		}
	})
	time.Sleep(time.Second * 2)

	t.Run("(4) переход к следующей песне", func(t *testing.T) {
		err = client.Next()
		if err != nil {
			t.Fatalf("failed to playing a next song: %v", err)
		}
	})
	time.Sleep(time.Second * 4)

	t.Run("(5) переход к предыдущей песне", func(t *testing.T) {
		err = client.Prev()
		if err != nil {
			t.Fatalf("failed to playing a previous song: %v", err)
		}
	})
	time.Sleep(time.Second * 2)

	t.Run("(6) добавление новой песни в конец плейлиста", func(t *testing.T) {
		song, err := client.AddSong(&Song{Description: "Lion in the Winter", Duration: 20})
		if err != nil {
			t.Fatalf("failed to adding a song: %v", err)
		}
		fmt.Println("добавлена песня:", song)
	})

}
