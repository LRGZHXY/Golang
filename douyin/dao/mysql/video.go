package mysql

import "database/sql"

func GetVideoList(page, limit int) ([]string, error) {
	//计算要跳过多少个数据
	offset := (page - 1) * limit

	sqlStr := "select video_id from videos order by id limit ? offset ?"

	var videoIdList []string
	if err := db.Select(&videoIdList, sqlStr, limit, offset); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return videoIdList, err
	}
	return videoIdList, nil
}
