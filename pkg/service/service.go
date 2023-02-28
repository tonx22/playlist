package service

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	Models "github.com/tonx22/playlist/pkg/models"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	trackCh = make(chan string)
	doneCh  = make(chan struct{})
)

type PlaylistService interface {
	Play(id int) error
	Pause() error
	Next() error
	Prev() error
	AddSong(song *Models.Song) (*Models.Song, error)
	Delete(id int) error
	GetPlaylist() (*[]Models.Song, error)
}

// начинает воспроизведение трека с указанным id, либо текущего трека если id = 0
// если id не указан и нет текущего трека - воспроизведение начинается с первого трека в плейлисте
func (svc *playlistService) Play(id int) error {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	if (svc.currentId > 0) && ((id == 0) || (id == svc.currentId)) {
		// продолжаем воспроизведение текущего трека
		if !svc.inProgress {
			select {
			case trackCh <- "play":
				svc.inProgress = true
			case <-time.After(10 * time.Second):
				return Models.ResponseError{ErrorDescr: "Playback timeout expired"}
			}
		}
	} else {
		// начинаем воспроизведение нового трека
		song, err := svc.GetTrack(id)
		if err != nil {
			return err
		}
		if svc.currentId > 0 {
			select {
			case trackCh <- "cancel":
				svc.currentId = 0
				svc.inProgress = false
			case <-time.After(10 * time.Second):
				return Models.ResponseError{ErrorDescr: "Playback timeout expired"}
			}
		}
		go playback(song)
		svc.currentId = song.Id
		svc.inProgress = true
	}
	return nil
}

// воспроизведение песни с возможностью приостановки/возобновления
func playback(song *Models.Song) {
	inProgress := true
	playbackTime := song.Duration
	for playbackTime > 0 {
		select {
		case command := <-trackCh:
			switch command {
			case "pause":
				inProgress = false
			case "play":
				inProgress = true
			case "cancel":
				return
			}
		case <-time.After(1 * time.Second):
			// очередная секунда воспроизведения
			if inProgress {
				log.Println(song.Description)
				playbackTime--
			}
		}
	}
	select {
	case doneCh <- struct{}{}:
	default:
	}
}

// поиск трека по id, либо первого в плейлисте если id = 0
func (svc *playlistService) GetTrack(id int) (*Models.Song, error) {
	var rows *sql.Rows
	var err error
	if id == 0 {
		rows, err = svc.DB.Query("select id, description, duration from playlist where prev = 0 limit 1")
	} else {
		rows, err = svc.DB.Query("select id, description, duration from playlist where id = $1 limit 1", id)
	}
	if err != nil {
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}
	defer rows.Close()

	var row_id, duration int
	var description string
	for rows.Next() {
		err := rows.Scan(&row_id, &description, &duration)
		if err != nil {
			return nil, Models.ResponseError{ErrorDescr: err.Error()}
		}
		return &Models.Song{Id: row_id, Description: description, Duration: duration}, nil
	}
	return nil, Models.ResponseError{ErrorDescr: "No data on request parameters", Status: http.StatusNotFound}
}

func (svc *playlistService) GetNextTrack(id int) (*Models.Song, error) {
	var rows *sql.Rows
	var err error
	rows, err = svc.DB.Query("select id, description, duration from playlist where prev = $1 limit 1", id)
	if err != nil {
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}
	defer rows.Close()

	var row_id, duration int
	var description string
	for rows.Next() {
		err := rows.Scan(&row_id, &description, &duration)
		if err != nil {
			return nil, Models.ResponseError{ErrorDescr: err.Error()}
		}
		return &Models.Song{Id: row_id, Description: description, Duration: duration}, nil
	}
	return nil, Models.ResponseError{ErrorDescr: "No data on request parameters", Status: http.StatusNotFound}
}

func (svc *playlistService) GetPrevTrack(id int) (*Models.Song, error) {
	var rows *sql.Rows
	var err error
	rows, err = svc.DB.Query("select id, description, duration from playlist where next = $1 limit 1", id)
	if err != nil {
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}
	defer rows.Close()

	var row_id, duration int
	var description string
	for rows.Next() {
		err := rows.Scan(&row_id, &description, &duration)
		if err != nil {
			return nil, Models.ResponseError{ErrorDescr: err.Error()}
		}
		return &Models.Song{Id: row_id, Description: description, Duration: duration}, nil
	}
	return nil, Models.ResponseError{ErrorDescr: "No data on request parameters", Status: http.StatusNotFound}
}

func (svc *playlistService) Pause() error {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	if svc.inProgress {
		select {
		case trackCh <- "pause":
			svc.inProgress = false
		case <-time.After(10 * time.Second):
			return Models.ResponseError{ErrorDescr: "Playback timeout expired"}
		}
	}
	return nil
}

func (svc *playlistService) Next() error {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	if svc.currentId > 0 {
		song, err := svc.GetNextTrack(svc.currentId)
		if err != nil {
			return Models.ResponseError{ErrorDescr: err.Error()}
		} else {
			select {
			case trackCh <- "cancel":
				svc.currentId = 0
				svc.inProgress = false
			case <-time.After(10 * time.Second):
				return Models.ResponseError{ErrorDescr: "Playback timeout expired"}
			}
			go playback(song)
			svc.currentId = song.Id
			svc.inProgress = true
		}
	} else {
		return Models.ResponseError{ErrorDescr: "Missing current track", Status: http.StatusNotFound}
	}
	return nil
}

func (svc *playlistService) Prev() error {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	if svc.currentId > 0 {
		song, err := svc.GetPrevTrack(svc.currentId)
		if err != nil {
			return Models.ResponseError{ErrorDescr: err.Error()}
		} else {
			select {
			case trackCh <- "cancel":
				svc.currentId = 0
				svc.inProgress = false
			case <-time.After(10 * time.Second):
				return Models.ResponseError{ErrorDescr: "Playback timeout expired"}
			}
			go playback(song)
			svc.currentId = song.Id
			svc.inProgress = true
		}
	} else {
		return Models.ResponseError{ErrorDescr: "Missing current track", Status: http.StatusNotFound}
	}
	return nil
}

func (svc *playlistService) AddSong(song *Models.Song) (*Models.Song, error) {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	ctx := context.Background()
	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}

	row := tx.QueryRow("WITH inserted AS (INSERT INTO playlist (description, duration, prev) VALUES "+
		"($1, $2, (SELECT coalesce(id, 0) FROM playlist WHERE next = 0 limit 1)) RETURNING id) SELECT id FROM inserted limit 1", song.Description, song.Duration)
	var id int
	err = row.Scan(&id)
	if err != nil {
		tx.Rollback()
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}

	_, err = tx.ExecContext(ctx, "UPDATE playlist SET next = $1 WHERE next = 0 AND id != $1", id)
	if err != nil {
		tx.Rollback()
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}
	song.Id = id
	return song, nil
}

func (svc *playlistService) Delete(id int) error {
	if id <= 0 {
		return Models.ResponseError{ErrorDescr: "Incorrect track id, must be > 0"}
	}
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	if (svc.currentId == id) && (svc.inProgress) {
		return Models.ResponseError{ErrorDescr: "Impossible to remove, the track is played"}
	}

	ctx := context.Background()
	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		return Models.ResponseError{ErrorDescr: err.Error()}
	}

	_, err = tx.ExecContext(ctx, "WITH deleted AS (SELECT prev, next FROM playlist WHERE id = $1 limit 1) UPDATE playlist SET "+
		"next = CASE WHEN next = $1 THEN (SELECT next FROM deleted) ELSE next END, "+
		"prev = CASE WHEN prev = $1 THEN (SELECT prev FROM deleted) ELSE prev END "+
		"WHERE (next = $1) OR (prev = $1)", id)
	if err != nil {
		tx.Rollback()
		return Models.ResponseError{ErrorDescr: err.Error()}
	}

	_, err = tx.ExecContext(ctx, "delete from playlist where id = $1", id)
	if err != nil {
		tx.Rollback()
		return Models.ResponseError{ErrorDescr: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return Models.ResponseError{ErrorDescr: err.Error()}
	}

	if svc.currentId == id {
		svc.currentId = 0
		svc.inProgress = false
	}
	return nil
}

func (svc *playlistService) GetPlaylist() (*[]Models.Song, error) {
	rows, err := svc.DB.Query("select * from playlist")
	if err != nil {
		return nil, Models.ResponseError{ErrorDescr: err.Error()}
	}
	defer rows.Close()

	var id, duration, prev, next int
	var description string
	r := make([]Models.Song, 0)
	for rows.Next() {
		err := rows.Scan(&id, &description, &duration, &prev, &next)
		if err != nil {
			return nil, Models.ResponseError{ErrorDescr: err.Error()}
		}
		r = append(r, Models.Song{Id: id, Description: description, Duration: duration, Prev: prev, Next: next})
	}
	return &r, nil
}

// реализует последовательное воспроизведение треков плейлиста
func (svc *playlistService) player() {
	for {
		select {
		case <-doneCh:
			// завершено воспроизведение текущего трека, начинаем следующий
			svc.mtx.Lock()
			if svc.currentId > 0 {
				song, err := svc.GetNextTrack(svc.currentId)
				if err != nil {
					svc.currentId = 0
					svc.inProgress = false
				} else {
					go playback(song)
					svc.currentId = song.Id
					svc.inProgress = true
				}
			}
			svc.mtx.Unlock()
		}
	}
}

type playlistService struct {
	mtx        sync.RWMutex
	currentId  int
	inProgress bool
	DB         *sql.DB
}

func NewPlaylistService(postgresUri string) (*playlistService, error) {
	time.Sleep(2 * time.Second)
	db, err := sql.Open("postgres", postgresUri)
	if err != nil {
		return nil, fmt.Errorf("Can't connect to postgresql: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Can't test ping to postgresql: %v", err)
	}
	svc := playlistService{DB: db}
	go svc.player()
	return &svc, nil
}

func Shutdown(s *playlistService) {
	_ = s.DB.Close()
}
