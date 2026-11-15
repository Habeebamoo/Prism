package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/middlewares"
	"github.com/Habeebamoo/Prism/internal/models"
	"github.com/Habeebamoo/Prism/internal/storage"
	"github.com/Habeebamoo/Prism/pkg/utils"
)

type VideoHandler struct {
	cfg  *configs.Config
	DB   *sql.DB
	B2   *storage.B2Client
}

type VideoRecord struct {
	VideoId  string
	UserId   string
	Title    string
	Status   string
}

func NewVideoHandler(cfg *configs.Config, db *sql.DB, B2 *storage.B2Client) *VideoHandler {
	return &VideoHandler{ cfg: cfg, DB: db, B2: B2 }
}

func (v *VideoHandler) CreateVideoRecord(ctx context.Context, record VideoRecord) error {
	_, err := v.DB.ExecContext(
		ctx,
		`INSERT INTO videos (id, user_id, title, status)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE
		 SET user_id = COALESCE(videos.user_id, EXCLUDED.user_id),
		     status = EXCLUDED.status`,
		record.VideoId,
		record.UserId,
		record.Title,
		record.Status,
	)

	return err
}

func (v *VideoHandler) UpdateVideoStatus(videoID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := v.DB.ExecContext(
		ctx,
		`UPDATE videos SET status = $1 WHERE id = $2`,
		status,
		videoID,
	)
	return err
}

func (v *VideoHandler) HandleListMyVideos(w http.ResponseWriter, r *http.Request) {
	userId, _ := middlewares.UserIDFromContext(r.Context())

	rows, err := v.DB.QueryContext(
		r.Context(),
		`SELECT id, user_id, status, title, created_at
		 FROM videos
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userId,
	)
	if err != nil {
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Failed to load videos" })
		return
	}
	defer rows.Close()

	videos := []models.Video{}

	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.Id, &v.UserId, &v.Status, &v.Title, &v.CreatedAt); err != nil {
			utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Failed to scan videos" })
			return
		}

		videos = append(videos, v)
	}

	if err := rows.Err(); err != nil {
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Error iterating video rows" })
		return
	}

	utils.JsonResponse(w, 200, utils.Payload{ Success: true, Data: videos })
}

func (v *VideoHandler) HandleDeleteVideo(w http.ResponseWriter, r *http.Request) {
	userId, _ := middlewares.UserIDFromContext(r.Context())

	videoId := r.PathValue("id")
	if videoId == "" {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Missing Video Id" })
		return
	}

	res, err := v.DB.ExecContext(
		r.Context(),
		`DELETE FROM videos WHERE id = $1 AND user_id = $2`,
		videoId, userId,
	)
	if err != nil {
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Delete Failed" })
		return
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		utils.JsonResponse(w, 404, utils.Payload{ Success: false, Message: "Video Not Found" })
		return
	} 

	// run in a seperate goroutine
	go v.DeleteFromB2(videoId)

	utils.JsonResponse(w, 200, utils.Payload{ Success: true, Message: "Video Deleted" })
}

func (v *VideoHandler) DeleteFromB2(videoID string) {
	if v.B2 == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := v.B2.DeleteVideoFolder(ctx, videoID); err != nil {
		log.Printf("[B2] failed to delete video %s: %v", videoID, err)
		return
	}

	log.Printf("[B2] deleted HLS folder for video %s", videoID)
}
