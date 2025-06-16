package middleware

import (
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/logic/domainservice"
	"github.com/gin-gonic/gin"
)

func AuthUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("go-mall-token")
		if len(token) != 40 {
			app.NewResponse(c).Error(errcode.ErrToken)
			c.Abort()
			return
		}
		tokenVerify, err := domainservice.NewUserDomainSvc(c).VerifyAuthToken(token)
		if err != nil {
			app.NewResponse(c).Error(errcode.ErrServer)
			c.Abort()
			return
		}
		if !tokenVerify.Approved {
			app.NewResponse(c).Error(errcode.ErrToken)
			c.Abort()
			return
		}
		c.Set("user_id", tokenVerify.UserId)
		c.Set("sessionId", tokenVerify.SessionId)
		c.Next()
	}
}
