package main

import (
	dbase "Users/sahithyakamalapadu/Desktop/Uptime/db"
	"Users/sahithyakamalapadu/Desktop/Uptime/handler"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

//UserModel for database

type info struct {
	done chan (bool)
	data chan (bool)
	//Crawltimeout  int
	Freq int
	
}

//var db *gorm.DB

func main() {

	handler.Connectdb()
	
	router := gin.Default()
	m := make(map[int]signal)
	
	
	router.POST("/urls/", handler.Post(m))
	router.GET("/urls/:id", handler.Getbyid())
    router.PATCH("/urls/:id", handler.Patch(m)) 
	router.POST("/urls/:id/activate", handler.Activate(m))
	router.POST("/urls/:id/deactivate", handler.Deactivate(m)) {
	router.DELETE("/urls/:id", handler.Deletebyid(m))

	router.Run()
	db.Close()
}
