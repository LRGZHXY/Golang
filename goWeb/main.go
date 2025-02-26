package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	//sqlx
	"github.com/jmoiron/sqlx"
	//mysql
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	//初始化一个gin路由
	r := gin.Default()
	//注册路由
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.Run()

	db, err := sqlx.Connect("mysql", "user:password@tcp(localhost:3306)/dbname")
}
