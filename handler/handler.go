package handler

import (
	"Users/sahithyakamalapadu/Desktop/Uptime/db"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	//mysql
	_ "github.com/go-sql-driver/mysql"
)

//Lock for mutex
var Lock sync.Mutex

type info struct {
	done chan (bool)
	data chan (bool)
	//Crawltimeout  int
	Freq int
	//Failthreshold int
}

var r = db.Database{}
var p = &r

//Post handler
func Post(m map[int]info) func(c *gin.Context) {
	return func(c *gin.Context) {
		url := c.PostForm("url")
		crawltime, _ := strconv.Atoi(c.PostForm("crawl_time"))
		freq, _ := strconv.Atoi(c.PostForm("frequency"))
		failthreshold, _ := strconv.Atoi(c.PostForm("fail_threshold"))

		user := db.UserModel{URL: url, Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}
		user.Stat = "active"
		//fmt.Printf("Here is the url of request: %s", url)
		Lock.Lock()
		urlid := int(r.Insert(&user))
		Lock.Unlock()
		m[urlid] = info{done: make(chan bool), data: make(chan bool), Freq: freq} //Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}
		go bgcheck(urlid, m, m[urlid].Freq, m[urlid].done, &user)

		c.JSON(http.StatusOK, gin.H{"response": " ", "ID": user, "Added into DB": true})
	}
}

//Getbyid handler
func Getbyid(m map[int]info) func(c *gin.Context) {
	return func(c *gin.Context) {
		var user db.UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))
		err := r.GetUrl(&user, urlid)
		if err != nil {
			c.AbortWithStatus(500)
		} else {
			c.JSON(http.StatusOK, gin.H{"response": r.First(&user, urlid)})
		}

	}
}

//Patch handler
func Patch(m map[int]info) func(c *gin.Context) {
	return func(c *gin.Context) {
		var user db.UserModel

		urlid, _ := (strconv.Atoi(c.Param("id")))
		crawltime, _ := strconv.Atoi(c.PostForm("crawl_time"))
		freq, _ := strconv.Atoi(c.PostForm("frequency"))
		failthreshold, _ := strconv.Atoi(c.PostForm("fail_threshold"))

		r.First(&user, urlid)

		if user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "500", "Info": "Url record doesn't exist"})
			return
		}
		Lock.Lock()
		url := r.PatchUpdate(&user, urlid, crawltime, freq, failthreshold)
		_ = url
		Lock.Unlock()
		m[urlid] = info{done: make(chan bool), data: make(chan bool), Freq: freq} //Crawltimeout: crawltime, Freq: freq, Failthreshold: failthreshold}
		m[urlid].data <- true
		c.JSON(http.StatusOK, gin.H{"response": " ", "url": user, "Updated into DB": true})
	}
}

//Activate handler
func Activate(m map[int]info) func(c *gin.Context) {
	return func(c *gin.Context) {
		var user db.UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))

		r.First(&user, urlid)
		if user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "500", "Info": "Url record doesn't exist"})
			return
		}
		Lock.Lock()
		err := r.Activate(urlid)
		Lock.Unlock()
		if err != nil {
			c.String(400, "Bad Request : URL is already Inctive")
		}
		c.JSON(200, gin.H{"response": "", "Info": "url is activated"})

		urlid = int(urlid)
		m[urlid] = info{done: make(chan bool), Freq: user.Freq}

	}
}

//Deactivate handler
func Deactivate(m map[int]info) func(c *gin.Context) {
	return func(c *gin.Context) {
		var user db.UserModel
		urlid, _ := strconv.Atoi(c.Param("id"))

		r.First(&user, urlid)
		if user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "500", "Info": "Url record doesn't exist"})
			return
		}
		Lock.Lock()
		err := r.Deactivate(urlid)
		Lock.Unlock()
		if err != nil {
			c.String(400, "Bad Request : URL is already Inctive")
		}
		Lock.Lock()
		m[int(urlid)].done <- true
		Lock.Unlock()
		c.String(200, "Success")

	}
}

//Deletebyid handler
func Deletebyid(m map[int]info) func(c *gin.Context) {
	return func(c *gin.Context) {
		urlid, _ := strconv.Atoi(c.Param("id"))
		err := r.Delete(urlid)
		_ = err
		m[int(urlid)].done <- true

		c.String(204, "Deleted Succefully")
	}
}

func check(s string, Crawltimeout int, user *db.UserModel, done chan (bool)) (stat string) {
	c2 := make(chan string)
	go func() {
		resp, err := http.Get(s)
		if err != nil || resp.StatusCode != 200 {
			r.Update(user, user.Failcount+1, "")

			if user.Failcount > user.Failthreshold {
				r.Deactivate(int(user.ID))
				done <- true
			}
			c2 <- "Not Found"
			return
		}
		if user.Stat == "Inactive" && resp.StatusCode == 200 {
			r.Update(user, 0, "Active")
		}
		c2 <- string(resp.StatusCode)
	}()

	select {
	case result := <-c2:
		return result

	case <-time.After(time.Duration(Crawltimeout) * time.Second):
		{
			r.Update(user, user.Failcount+1, "")
			return "GiveUp"
		}
	}
}

func bgcheck(id int, m map[int]info, fq int, done chan (bool), user *db.UserModel) {
	freq := m[id].Freq
	crawl := user.Crawltimeout
	ticker := time.NewTicker(time.Second * time.Duration(freq))
	for range ticker.C {

		res := check(user.URL, crawl, user, done)

		d, _ := <-m[id].data

		if d {
			freq = m[id].Freq
			crawl = user.Crawltimeout
			ticker = time.NewTicker(time.Second * time.Duration(freq))
		}
		v, _ := <-done
		if v {
			ticker.Stop()
			return
		}
		if res == "Giveup" {
			ticker.Stop()
			return
		}
	}
}

//Connectdb creates connection to db
func Connectdb() {
	err := r.CreateConnection()
	_ = err
}
func Closedb(){
	r.CloseDB()
}
