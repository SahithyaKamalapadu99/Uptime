package db

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type Suite struct {
	suite.Suite
	DB    *gorm.DB
	mock  sqlmock.Sqlmock
	repo  Repository
	model UserModel
}

func (s *Suite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.DB, err = gorm.Open("postgres", db)
	require.NoError(s.T(), err)

	s.DB.LogMode(true)

	s.repo = CreateRepository(s.DB)
}
func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}
func (s *Suite) Test_repository_GetUrl() {
	var (
		id            = 1
		Url           = "abc.com"
		Crawltime     = 10
		freq          = 20
		failthreshold = 3
		status        = "active"
		failcount     = 0
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "user_models" WHERE (id = $1)`)).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawltimeout", "freq", "failthreshold", "stat", "failcount"}).
			AddRow(id, Url, Crawltime, freq, failthreshold, status, failcount))
	var user UserModel
	err := s.repo.GetUrl(&user, id)

	require.NoError(s.T(), err)
	//require.Nil(s.T(), deep.Equal(&UserModel{ID: uint(id), URL: Url, Crawltimeout: Crawltime, Freq: freq, Failthreshold: failthreshold, Stat: status, Failcount: failcount}, res))
}
