package data

import (
	"douyin/dao/mysql"
	"douyin/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetDataByID(c *gin.Context) {
	videoID := c.Param("id")

	aweme, err := mysql.GetDataByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Video not found: " + err.Error(),
		})
		return
	}

	res := models.ResponseData{}
	res.Data.VideoID = videoID
	res.Data.AwemeDetail = aweme
	res.Data.Success = true

	c.JSON(http.StatusOK, res)
}
