package videoids

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const dataDir3 = "E:\\GoPractice\\douyin\\extra\\videoids"

// 定义一个映射，将 id 映射到对应的 JSON 文件
var videoIdMap = map[string]string{
	"7482354135765880091": "video1.json",
	"7477984552787397946": "video2.json",
	"7474583086403980604": "video3.json",
	"7461981146180537641": "video4.json",
	"7481509204755369270": "video5.json",
	"7480744910195363124": "video6.json",
}

// GetVideoId 根据 URL 中的 id 返回对应的视频文件内容
func GetVideoId(c *gin.Context) {
	// 从 URL 获取 id 参数
	id := c.Param("id")

	// 映射 id 到具体的 json 文件名
	fileName, ok := videoIdMap[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("未找到 %s 对应的数据文件", id),
		})
		return
	}

	// 构建文件路径
	filePath := filepath.Join(dataDir3, fileName)

	// 打开对应的json文件
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
