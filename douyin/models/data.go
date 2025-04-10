package models

type AwemeDetail struct {
	AwemeID    string `json:"aweme_id"`
	Desc       string `json:"desc"`
	Author     Author `json:"author"`
	Music      Music  `json:"music"`
	Statistics Stats  `json:"statistics"`
	Videos     Videos `json:"videos"`
}

type Author struct {
	UID      string `json:"uid"`
	SecUID   string `json:"sec_uid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar_url"`
}

type Music struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	CoverURL string `json:"cover_url"`
}

type Stats struct {
	DiggCount    int `json:"digg_count"`
	CommentCount int `json:"comment_count"`
	ShareCount   int `json:"share_count"`
}

type PlayAddr struct {
	URLList []string `json:"url_list"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
}

type Videos struct {
	PlayAddr PlayAddr `json:"play_addr"`
	Width    int      `json:"width"`
	Height   int      `json:"height"`
}

// 接收请求体
type RequestData struct {
	Data struct {
		AwemeDetail AwemeDetail `json:"aweme_detail"`
	} `json:"data"`
}

// 返回响应体
type ResponseData struct {
	Data struct {
		VideoID     string      `json:"videoID"`
		AwemeDetail AwemeDetail `json:"aweme_detail"`
		Success     bool        `json:"success"`
	} `json:"data"`
}
