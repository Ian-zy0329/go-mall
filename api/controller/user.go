package controller

import (
	"errors"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/logic/appservice"
	"github.com/gin-gonic/gin"
	"strconv"
)

func RegisterUser(c *gin.Context) {
	userRequest := new(request.UserRegister)
	if err := c.ShouldBind(userRequest); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	if !util.PasswordComplexityVerify(userRequest.Password) {
		logger.New(c).Warn("RegisterUserError", "err", "密码复杂度不满足", "password", userRequest.Password)
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.UserRegister(userRequest)
	if err != nil {
		if errors.Is(err, errcode.ErrUserNameOccupied) {
			app.NewResponse(c).Error(errcode.ErrUserNameOccupied)
		} else {
			appErr := err.(*errcode.AppError)
			app.NewResponse(c).Error(appErr)
		}
		return
	}
	app.NewResponse(c).SuccessOk()
	return
}

func LoginUser(c *gin.Context) {
	loginRequest := new(request.UserLogin)
	if err := c.ShouldBindJSON(&loginRequest.Body); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	if err := c.ShouldBindHeader(&loginRequest.Header); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	token, err := userSvc.UserLogin(loginRequest)
	if err != nil {
		if errors.Is(err, errcode.ErrUserNotRight) {
			app.NewResponse(c).Error(errcode.ErrUserNotRight)
		} else if errors.Is(err, errcode.ErrUserInvalid) {
			app.NewResponse(c).Error(errcode.ErrUserNotRight)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		logger.New(c).Error("LoginUserError", "err", err)
		return
	}
	app.NewResponse(c).Success(token)
	return
}

func LogoutUser(c *gin.Context) {
	userId := c.GetInt64("userId")
	platform := c.GetString("platform")
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.UserLogout(userId, platform)
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		return
	}
	app.NewResponse(c).SuccessOk()
}

func PasswordResetApply(c *gin.Context) {
	request := new(request.PasswordResetApply)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	reply, err := userSvc.PasswordResetApply(request)
	if err != nil {
		if errors.Is(err, errcode.ErrUserNotRight) {
			app.NewResponse(c).Error(errcode.ErrUserNotRight)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).Success(reply)
}

func PasswordReset(c *gin.Context) {
	request := new(request.PasswordReset)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	if !util.PasswordComplexityVerify(request.Password) {
		logger.New(c).Warn("RegisterUserError", "err", "密码复杂度不满足", "password", request.Password)
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(errors.New("密码复杂度不达标")))
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.PasswordReset(request)
	if err != nil {
		if errors.Is(err, errcode.ErrParams) {
			app.NewResponse(c).Error(errcode.ErrParams)
		} else if errors.Is(err, errcode.ErrUserInvalid) {
			app.NewResponse(c).Error(errcode.ErrUserInvalid)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer)
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}

func UserInfo(c *gin.Context) {
	userId := c.GetInt64("userId")
	userSvc := appservice.NewUserAppSvc(c)
	userInfoReply := userSvc.UserInfo(userId)
	if userInfoReply == nil {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}
	app.NewResponse(c).Success(userInfoReply)
}

func UpdateUserInfo(c *gin.Context) {
	request := new(request.UserInfoUpdate)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.UserInfoUpdate(request, c.GetInt64("userId"))
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		return
	}
	app.NewResponse(c).SuccessOk()
}

func AddUserAddress(c *gin.Context) {
	request := new(request.UserAddress)
	if err := c.ShouldBindJSON(request); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.AddUserAddress(request, c.GetInt64("userId"))
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		return
	}
	app.NewResponse(c).SuccessOk()
}

func GetUserAddresses(c *gin.Context) {
	userSvc := appservice.NewUserAppSvc(c)
	replyData, err := userSvc.GetUserAddresses(c.GetInt64("userId"))
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		return
	}
	app.NewResponse(c).Success(replyData)
}

func UpdateUserAddress(c *gin.Context) {
	addressId, _ := strconv.ParseInt(c.Param("address_id"), 10, 64)
	if addressId <= 0 {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}
	requestData := new(request.UserAddress)
	if err := c.ShouldBind(requestData); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.ModifyUserAddress(requestData, c.GetInt64("userId"), addressId)
	if err != nil {
		if errors.Is(err, errcode.ErrParams) {
			app.NewResponse(c).Error(errcode.ErrParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer)
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}

func GetSingleAddress(c *gin.Context) {
	addressId, _ := strconv.ParseInt(c.Param("address_id"), 10, 64)
	if addressId <= 0 {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}

	userSvc := appservice.NewUserAppSvc(c)
	singleAddress, err := userSvc.GetSingleAddress(addressId)
	if err != nil {
		if errors.Is(err, errcode.ErrParams) {
			app.NewResponse(c).Error(errcode.ErrParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer)
		}
		return
	}
	app.NewResponse(c).Success(singleAddress)
}

func DeleteUserAddress(c *gin.Context) {
	addressId, _ := strconv.ParseInt(c.Param("address_id"), 10, 64)
	if addressId <= 0 {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	err := userSvc.DeleteUserAddress(c.GetInt64("userId"), addressId)
	if err != nil {
		if errors.Is(err, errcode.ErrParams) {
			app.NewResponse(c).Error(errcode.ErrParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer)
		}
		return
	}
	app.NewResponse(c).SuccessOk()
}
