package models

type Video struct {
	ID      int64  `db:"id"`
	VideoID string `db:"video_id"`
}
