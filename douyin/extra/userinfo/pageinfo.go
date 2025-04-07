package userinfo

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const dataDir2 = "E:\\GoPractice\\douyin\\extra\\userinfo"

// 用户 ID 与本地 JSON 文件名的映射表
var pageFileMap = map[string]string{
	"MS4wLjABAAAANXSltcLCzDGmdNFI2Q_QixVTr67NiYzjKOIP5s03CAE":                      "user1.1.json",
	"MS4wLjABAAAAW9FWcqS7RdQAWPd2AA5fL_ilmqsIFUCQ_Iym6Yh9_cUa6ZRqVLjVQSUjlHrfXY1Y": "user2.1.json",
}

func FetchUserPageVideos(c *gin.Context) {
	// 获取 URL 查询参数
	secUserID := c.Query("sec_user_id")
	maxCursor := c.Query("max_cursor")
	counts := c.Query("counts")

	// 校验参数是否缺失
	if secUserID == "" || maxCursor == "" || counts == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要参数：sec_user_id, max_cursor 或 counts",
		})
		return
	}

	// 映射 sec_user_id 到具体的 JSON 文件名
	filename, ok := pageFileMap[secUserID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("未找到 sec_user_id %s 对应的数据文件", secUserID),
		})
		return
	}

	// 构建文件路径
	jsonFilePath := filepath.Join(dataDir2, filename)

	// 尝试打开 JSON 文件
	file, err := os.Open(jsonFilePath)
	if err != nil {
		log.Printf("打开文件失败: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "找不到该用户的数据文件111",
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
