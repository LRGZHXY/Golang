package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
)

func CreatePost(p *models.Post) (err error) {
	//1.生成post id
	p.ID = snowflake.GenID()
	//2.保存到数据库
	return mysql.CreatePost(p)
	//3.返回
}

// 根据帖子id查询帖子详情数据
func GetPostById(pid int64) (data *models.Post, err error) {
	return mysql.GetPostById(pid)
}
