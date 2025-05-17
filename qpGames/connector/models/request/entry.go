package request

/*{
	token: 'e26ad95f109145...'
	userInfo: { nickname:'码神710065',avatar:'Common/ead_icon_default',sex:1 }
}*/

type EntryReq struct {
	Token    string   `json:"token"`
	UserInfo UserInfo `json:"userInfo"`
}

type UserInfo struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Sex      int    `json:"sex"`
}
