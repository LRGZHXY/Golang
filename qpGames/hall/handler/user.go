package handler

import (
	"common"
	"common/biz"
	"common/logs"
	"core/repo"
	"core/service"
	"encoding/json"
	"framework/remote"
	"hall/models/request"
	"hall/models/response"
)

type UserHandler struct {
	userService *service.UserService
}

// UpdateUserAddress 更新用户地址
func (h *UserHandler) UpdateUserAddress(session *remote.Session, msg []byte) any {
	logs.Info("UpdateUserAddress msg:%v", string(msg))
	var req request.UpdateUserAddressReq
	if err := json.Unmarshal(msg, &req); err != nil {
		return common.F(biz.RequestDataError)
	}
	err := h.userService.UpdateUserAddressByUid(session.GetUid(), req)
	if err != nil {
		return common.F(biz.SqlError)
	}
	res := response.UpdateUserAddressRes{}
	res.Code = biz.OK
	res.UpdateUserData = req
	return res
}

func NewUserHandler(r *repo.Manager) *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(r),
	}
}
