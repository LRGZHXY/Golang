package data

import (
	"douyin/dao/mysql"
	"douyin/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func TranData(c *gin.Context) {
	var req models.RequestData

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON",
		})
		return
	}

	videoID := uuid.New().String()
	aweme := req.Data.AwemeDetail

	// 存进数据库
	if err := mysql.AddData(videoID, aweme); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to insert data: " + err.Error(),
		})
		return
	}

	var res models.ResponseData
	res.Data.VideoID = videoID
	res.Data.AwemeDetail = aweme
	res.Data.Success = true

	c.JSON(http.StatusOK, res)
}
