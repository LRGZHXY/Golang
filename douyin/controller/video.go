package controller

import (
	"douyin/logic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// GetVideoHandler 获取视频id
func GetVideoHandler(c *gin.Context) {
	//1.获取参数和请求校验
	//从url中获取page的值
	pageStr := c.Param("page")
	page, err := strconv.Atoi(pageStr) //将page转化为整数
	if err != nil || page < 1 {
		page = 1
	}
	//2.业务处理
	videoIDs, err := logic.GetVideoList(page)
	if err != nil {
		zap.L().Error("logic.GetVideoList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//3.返回响应
	ResponseSuccess(c, videoIDs)
}
