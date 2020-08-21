package db

import (
	"errors"
	"fmt"
	"sync"

	//mysql
	_ "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
)

/*// Model is gorm.Model definition
type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}*/

//UserModel...
type UserModel struct {
	ID            uint `gorm:"primary_key"`
	URL           string
	Crawltimeout  int
	Freq          int
	Failthreshold int
	Stat          string
	Failcount     int
}

//Repository interface has all the methods....
type Repository interface {
	CreateConnection() error
	Insert(user *UserModel) (id uint)
	GetUrl(user *UserModel, s int) (err error)
	PatchUpdate(user *UserModel, id int, crawl int, freq int, failthresh int) (url string)
	Update(user *UserModel, f int, s string)
	Delete(id int) error
	Activate(id int) error
	Deactivate(id int) error
	First(user *UserModel, id int) *UserModel
	CloseDB()
}

//Database Model...
type Database struct {
	DB *gorm.DB
}

//Lock is...
var Lock sync.Mutex

//CreateConnection ...
func (r *Database) CreateConnection() error {
	var err error
	r.DB, err = gorm.Open("mysql", "root:root1234@tcp(localhost:3306)/urls?charset=utf8&parseTime=True")
	if err != nil {
		fmt.Println("Connection Failed to Open")
		return err
	}

	fmt.Println("Connection Established")
	r.DB.AutoMigrate(&UserModel{})
	return nil
}

//Insert URL
func (r *Database) Insert(user *UserModel) (id uint) {

	r.DB.Create(&user)
	return user.ID
}

//GetUrl by id
func (r *Database) GetUrl(user *UserModel, s int) (err error) {

	if err := r.DB.Where("id = ?", s).First(&user).Error; err != nil {
		return err
	}
	return nil
}

//PatchUpdate for url
func (r *Database) PatchUpdate(user *UserModel, id int, crawl int, freq int, failthresh int) (url string) {

	r.DB.Model(&user).Updates(UserModel{Crawltimeout: crawl, Freq: freq, Failthreshold: failthresh, Failcount: 0})
	return user.URL
}

//Update record
func (r *Database) Update(user *UserModel, f int, s string) {
	r.DB.Model(&user).Update(UserModel{Failcount: f, Stat: s})
}

//Delete url
func (r *Database) Delete(id int) error {

	var user UserModel
	r.DB.First(&user, id)

	r.DB.Where("id=?", id).Delete(&user)
	return nil
}

//Activate url
func (r *Database) Activate(id int) error {
	var user UserModel
	r.DB.First(&user, id)

	if user.Stat == "Active" {
		return errors.New("Bad Request : URL is already Inctive")
	}
	r.DB.Model(&user).Updates(UserModel{Stat: "Active", Failcount: 0})
	return nil
}

//Deactivate url
func (r *Database) Deactivate(id int) error {
	var user UserModel
	r.DB.First(&user, id)

	if user.Stat == "Inactive" {
		return errors.New("Bad Request : URL is already Inctive")
	}

	r.DB.Model(&user).Update("Status", "Inactive")
	return nil
}

//First makes the query
func (r *Database) First(user *UserModel, id int) *UserModel {
	//var user UserModel
	r.DB.Find(&user, id)
	return user
}

//CloseDB connection
func (r *Database) CloseDB() {
	//c := Database{}
	r.DB.Close()
}

//CreateRepository tests the interface
func CreateRepository(db *gorm.DB) Repository {
	return &Database{
		DB: db,
	}
}
