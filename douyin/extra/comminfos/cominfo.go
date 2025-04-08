package videoid

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const dataDir4 = "E:\\GoPractice\\douyin\\extra\\comminfo"

// 定义一个映射，将 id 映射到对应的 JSON 文件
var commMap = map[string]string{
	"7482354135765880091": "com1.json",
	"7477984552787397946": "com2.json",
	"7474583086403980604": "com3.json",
	"7461981146180537641": "com4.json",
	"7481509204755369270": "com5.json",
	"7480744910195363124": "com6.json",
}

func GetComInfo(c *gin.Context) {
	// 获取 URL 查询参数
	awemeID := c.Query("aweme_id")
	cursor := c.Query("cursor")
	counts := c.Query("counts")

	// 校验参数是否缺失
	if awemeID == "" || cursor == "" || counts == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要参数：aweme_id, cursor 或 counts",
		})
		return
	}

	// 映射 id 到具体的 JSON 文件名
	fileName, ok := commMap[awemeID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("未找到 %s 对应的数据文件", awemeID),
		})
		return
	}

	// 构建文件路径
	filePath := filepath.Join(dataDir4, fileName)

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
