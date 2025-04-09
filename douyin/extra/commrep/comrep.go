package commrep

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const dataDir5 = "E:\\GoPractice\\douyin\\extra\\commrep"

// 定义一个映射，将 id 映射到对应的 JSON 文件
var commrepMap = map[string]string{
	"7482662504204927744": "rep1.json",
	"7482501495717315343": "rep2.json",
	"7482659705413518120": "rep3.json",
	"7487832192375292731": "rep4.json",
}

func FetchVideoCommentReplies(c *gin.Context) {
	// 获取 URL 查询参数
	itemID := c.Query("item_id")
	commentsID := c.Query("comment_id")
	cursor := c.Query("cursor")
	count := c.Query("count")

	// 校验参数是否缺失
	if itemID == "" || commentsID == "" || cursor == "" || count == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要参数：item_id,comment_id, cursor 或 counts",
		})
		return
	}

	// 映射 id 到具体的 JSON 文件名
	fileName, ok := commrepMap[commentsID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("未找到 %s 对应的数据文件", commentsID),
		})
		return
	}

	// 构建文件路径
	filePath := filepath.Join(dataDir5, fileName)

	// 尝试打开 JSON 文件
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("打开文件失败: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "找不到该文件",
		})
		return
	}
	defer file.Close()

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("读取文件内容失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器读取数据失败",
		})
		return
	}

	// 返回 JSON 数据
	c.Data(http.StatusOK, "application/json", data)
}
