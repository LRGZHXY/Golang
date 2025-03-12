package controllers

import (
	"bluebell/logic"
	"bluebell/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 投票
/*type VoteData struct {
	//UserID 从请求中获取当前的用户
	PostID    int64 `json:"post_id,string"`   //帖子id
	Direction int   `json:"direction,string"` //赞成票（1）还是反对票（-1）
}*/

func PostVoteController(c *gin.Context) {
	//参数校验
	p := new(models.ParamVoteData)
	if err := c.ShouldBind(&p); err != nil {
		errs, ok := err.(validator.ValidationErrors) //类型断言
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		errData := removeTopStruct(errs.Translate(trans)) //翻译并去除错误提示中的结构体标识
		ResponseErrorWithMsg(c, CodeInvalidParam, errData)
		return
	}
	//获取当前请求用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	//具体投票的业务逻辑
	logic.VoteForPost(userID, p)
	ResponseSuccess(c, nil)
}
