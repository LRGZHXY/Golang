package logic

import "douyin/dao/mysql"

func GetVideoList(page int) ([]string, error) {
	limit := 5
	return mysql.GetVideoList(page, limit)
}
