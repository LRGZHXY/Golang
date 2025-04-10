package mysql

import (
	"database/sql"
	"douyin/models"
	"errors"
)

func GetDataByID(videoID string) (models.AwemeDetail, error) {
	var aweme models.AwemeDetail
	var videoURL string
	var playWidth, playHeight int

	sqlStr := `SELECT
		aweme_id, description, author_uid, author_sec_uid, author_nickname, author_avatar,
		music_title, music_author, music_cover,
		digg_count, comment_count, share_count,
		video_url, video_width, video_height
	FROM aweme_videos WHERE id = ?`

	row := db.QueryRow(sqlStr, videoID)

	//Scan()将查询结果绑定到对应的变量中
	err := row.Scan(
		&aweme.AwemeID,
		&aweme.Desc,
		&aweme.Author.UID,
		&aweme.Author.SecUID,
		&aweme.Author.Nickname,
		&aweme.Author.Avatar,
		&aweme.Music.Title,
		&aweme.Music.Author,
		&aweme.Music.CoverURL,
		&aweme.Statistics.DiggCount,
		&aweme.Statistics.CommentCount,
		&aweme.Statistics.ShareCount,
		&videoURL,
		&playWidth,
		&playHeight,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return aweme, errors.New("no record found")
		}
		return aweme, err
	}

	// 组装视频播放地址结构
	aweme.Videos = models.Videos{
		PlayAddr: models.PlayAddr{
			URLList: []string{videoURL},
			Width:   playWidth,
			Height:  playHeight,
		},
		Width:  playWidth,
		Height: playHeight,
	}

	return aweme, nil
}
