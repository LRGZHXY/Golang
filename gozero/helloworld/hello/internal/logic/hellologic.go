package logic

import (
	"context"

	"GoPractice/gozero/helloworld/hello/internal/svc"
	"GoPractice/gozero/helloworld/hello/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HelloLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHelloLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HelloLogic {
	return &HelloLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HelloLogic) Hello(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
