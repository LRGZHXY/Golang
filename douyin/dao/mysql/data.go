package mysql

import (
	"douyin/models"
)

func AddData(videoID string, aweme models.AwemeDetail) error {
	videoURL := ""
	if len(aweme.Videos.PlayAddr.URLList) > 0 {
		videoURL = aweme.Videos.PlayAddr.URLList[0]
	}

	sqlStr := `INSERT INTO aweme_videos (
		id, aweme_id, description, author_uid, author_sec_uid, author_nickname, author_avatar,
		music_title, music_author, music_cover,
		digg_count, comment_count, share_count,
		video_url, video_width, video_height
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(sqlStr,
		videoID,
		aweme.AwemeID,
		aweme.Desc,
		aweme.Author.UID,
		aweme.Author.SecUID,
		aweme.Author.Nickname,
		aweme.Author.Avatar,
		aweme.Music.Title,
		aweme.Music.Author,
		aweme.Music.CoverURL,
		aweme.Statistics.DiggCount,
		aweme.Statistics.CommentCount,
		aweme.Statistics.ShareCount,
		videoURL,
		aweme.Videos.Width,
		aweme.Videos.Height,
	)

	return err
}
