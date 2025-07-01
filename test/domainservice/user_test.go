package domainservice

import (
	"github.com/Ian-zy0329/go-mall/common/util"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	SuppressConsoleStatistics()
	result := m.Run()
	PrintConsoleStatistics()
	os.Exit(result)
}

func TestPasswordComplexityVerify(t *testing.T) {
	Convey("Given a simple password", t, func() {
		password := "123456"
		Convey("when run it for password complexity checking", func() {
			result := util.PasswordComplexityVerify(password)
			Convey("Then the checking result should be false", func() {
				So(result, ShouldBeFalse)
			})
		})
	})
	Convey("Given a complex password", t, func() {
		password := "123456aA!"
		Convey("when run it for password complexity checking", func() {
			result := util.PasswordComplexityVerify(password)
			Convey("Then the checking result should be true", func() {
				So(result, ShouldBeTrue)
			})
		})
	})
}
