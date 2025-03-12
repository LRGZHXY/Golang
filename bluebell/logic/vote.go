package logic

import "bluebell/models"

//投票功能

/*投票的几种情况：
direction=1：
	1.之前没有投过票，现在投赞成票
	2.之前投反对票，现在改为赞成票
direction=0：
	1.之前投赞成票，现在取消投票
	2.之前投反对票，现在取消投票
direction=-1：
	1.之前没有投过票，现在投反对票
	2.之前投赞成票，现在改为反对票

投票的限制：
每个帖子自发表之日起一个星期之内允许用户投票，超过一个星期不允许投票
	1.到期之后将redis中保存的赞成票及反对票存储到mysql表中
	2.到期之后删除 KeyPostVotedZSetPF
*/

// VoteForPost 为帖子投票的函数
func VoteForPost(userID int64, p *models.ParamVoteData) {

}
