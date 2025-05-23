package controllers

import (
	"bluebell/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//跟社区相关

func CommunityHandler(c *gin.Context) {
	//查询到所有的社区(community_id,community_name)以列表的形式返回
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy) //不轻易把服务端报错暴漏给外面
		return
	}
	ResponseSuccess(c, data)
}

// CommunityDetailHandler 社区分类详情
func CommunityDetailHandler(c *gin.Context) {
	//1.获取社区id
	idStr := c.Param("id")                     //获取URL参数
	id, err := strconv.ParseInt(idStr, 10, 64) //将 idStr 字符串按照 10 进制解析成一个 int64 类型的整数
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	//2.根据id获取视频详情
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy) //不轻易把服务端报错暴漏给外面
		return
	}
	ResponseSuccess(c, data)
}
