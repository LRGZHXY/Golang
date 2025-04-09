package logic

import "douyin/dao/mysql"

func GetVideoList(page int, limit int) ([]string, error) {
	return mysql.GetVideoList(page, limit)
}
