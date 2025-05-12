package api

import (
	"common/logs"
	"common/rpc"
	"context"
	"github.com/gin-gonic/gin"
	"user/pb"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) Register(ctx *gin.Context) {
	response, err := rpc.UserClient.Register(context.TODO(), &pb.RegisterParams{})
	if err != nil {

	}
	uid := response.Uid
	logs.Info("uid:%s", uid)
}
