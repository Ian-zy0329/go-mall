package dao

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Ian-zy0329/go-mall/common/util"
	dao2 "github.com/Ian-zy0329/go-mall/dal/dao"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"regexp"
	"testing"
	"time"
)

var (
	mock sqlmock.Sqlmock
	err  error
	db   *sql.DB
)

func TestMain(m *testing.M) {
	db, mock, err = sqlmock.New()
	if err != nil {
		panic(err)
	}
	dbMasterConn, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
		DefaultStringSize:         0,
	}))
	dbSlaveConn, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
		DefaultStringSize:         0,
	}))
	dao2.SetDBMasterConn(dbMasterConn)
	dao2.SetDBSlaveConn(dbSlaveConn)
	os.Exit(m.Run())
}

func TestUserDao_CreateUser(t *testing.T) {
	type fields struct {
		ctx context.Context
	}
	userInfo := &do.UserBaseInfo{
		NickName:  "Slang",
		LoginName: "slang@go-mall.com",
		Verified:  "0",
		Avatar:    "",
		Slogan:    "happy!",
		IsBlocked: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	passwordHash, _ := util.BcryptPassword("123456")
	userIsDel := 0

	ud := dao2.NewUserDao(context.TODO())
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WithArgs(userInfo.NickName, userInfo.LoginName, passwordHash, userInfo.Verified, userInfo.Avatar,
			userInfo.Slogan, userIsDel, userInfo.IsBlocked, userInfo.CreatedAt, userInfo.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	userObj, err := ud.CreateUser(userInfo, passwordHash)
	assert.Nil(t, err)
	assert.Equal(t, userInfo.LoginName, userObj.LoginName)
}
