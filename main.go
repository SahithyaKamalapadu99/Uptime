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
	//data chan (int)
	//Crawltimeout  int
	Freq int
	//Failthreshold int
}

var db *gorm.DB

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

	/*router.POST("/urls/", func(c *gin.Context) {
		url := c.PostForm("url")
		crawltime, _ := strconv.Atoi(c.PostForm("crawl_time"))
		freq, _ := strconv.Atoi(c.PostForm("frequency"))
		failthreshold, _ := strconv.Atoi(c.PostForm("fail_threshold"))
		user := UserModel{URL: url, Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}
		user.Stat = "active"
		//fmt.Printf("Here is the url of request: %s", url)
		urlid := int(Insert(&user))

		m[urlid] = signal{done: make(chan bool), Freq: freq} //Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}

		go bgcheck(url, m[urlid].Freq, m[urlid].done, &user)

		c.JSON(http.StatusOK, gin.H{"response": " ", "ID": user, "Added into DB": true})
	})

	router.GET("/urls/:id", func(c *gin.Context) {
		var user UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))
		err := GetUrls(&user, urlid)
		if err != nil {
			c.AbortWithStatus(500)
		} else {
			//fmt.Println(db.First(&user, urlid))

			c.JSON(http.StatusOK, gin.H{"response": db.First(&user, urlid)})
		}

	})

	router.PATCH("/urls/:id", func(c *gin.Context) {
		var user UserModel
		urlid, _ := (strconv.Atoi(c.Param("id")))
		crawltime, _ := strconv.Atoi(c.PostForm("crawl_time"))
		freq, _ := strconv.Atoi(c.PostForm("frequency"))
		failthreshold, _ := strconv.Atoi(c.PostForm("fail_threshold"))

		db.First(&user, urlid)

		if user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "500", "Info": "Url record doesn't exist"})
			return
		}
		url := PatchUpdate(&user, urlid, crawltime, freq, failthreshold)

		m[urlid] = signal{done: make(chan bool), Freq: freq} //Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}

		c.JSON(http.StatusOK, gin.H{"response": " ", "url": url, "Updated into DB": true})

	})
	router.POST("/urls/:id/activate", func(c *gin.Context) {
		var user UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))
		db.First(&user, urlid)

		if user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "500", "Info": "Url record doesn't exist"})
			return
		}
		if user.Stat == "Active" {
			c.String(400, "Bad Request : URL is already Active")
		}
		db.Model(&user).Updates(UserModel{Stat: "Active", Failcount: 0})
		urlid = int(urlid)
		m[urlid] = signal{done: make(chan bool), Freq: user.Freq} //Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}

	})
	router.POST("/urls/:id/deactivate", func(c *gin.Context) {
		var user UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))
		db.First(&user, urlid)

		if user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "500", "Info": "Url record doesn't exist"})
			return
		}
		if user.Stat == "Inactive" {
			c.String(400, "Bad Request : URL is already Inctive")
		}
		m[int(urlid)].done <- true

		db.Model(&user).Update("Status", "Inactive")

	})
	router.DELETE("/urls/:id", func(c *gin.Context) {
		var user UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))
		//db.First(&user, urlid)
		m[int(urlid)].done <- true
		db.Where("id=?", urlid).Delete(&user)
	})
	

}

//check the URLs
func check(s string, user *UserModel, done chan (bool)) (stat string) {
	c2 := make(chan string)
	go func() {
		resp, err := http.Get(s)
		if err != nil || resp.StatusCode != 200 {
			db.Model(&user).Update("failcount", user.Failcount+1)
			if user.Failcount > user.Failthreshold {
				db.Model(&user).Update("Stat", "Inactive")
				done <- true
			}
			c2 <- "Not Found"
			return
		}
		if user.Stat == "Inactive" && resp.StatusCode == 200 {
			db.Model(&user).Update("Stat", "Active")
		}
		c2 <- string(resp.StatusCode)
	}()

	select {
	case result := <-c2:
		return result
	case <-time.After(time.Duration(user.Crawltimeout) * time.Second):
		{
			db.Model(&user).Update("failcount", user.Failcount+1)
			return "GiveUp"
		}
	}
}

func bgcheck(s string, fq int, done chan (bool), user *UserModel) {
	ticker := time.NewTicker(time.Second * time.Duration(fq))
	for range ticker.C {

		res := check(s, user, done)

		v, _ := <-done
		if v {
			ticker.Stop()
		}

		if res == "Giveup" {
			ticker.Stop()
		}
	}
}

//GetUrls Info
func GetUrls(user *UserModel, s int) (err error) {

	if err := db.Where("id = ?", s).First(&user).Error; err != nil {
		return err
	}
	return nil
}

//Insert data into DB
func Insert(user *UserModel) (id uint) {
	//user := UserModel{URL: url, Crawltimeout: ct, Freq: fq, Failthreshold: ft, Status: true}

	db.Create(&user)

	return user.ID

}

//PatchUpdate updates data for Patch request
func PatchUpdate(user *UserModel, id int, crawl int, freq int, failthresh int) (url string) {

	db.Model(&user).Updates(UserModel{Crawltimeout: crawl, Freq: freq, Failthreshold: failthresh, Failcount: 0})

	return user.URL
}
*/
router.Run()
	db.Close()
}