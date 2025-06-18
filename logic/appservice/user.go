package appservice

import (
	"context"
	"errors"
	"github.com/Ian-zy0329/go-mall/api/reply"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/Ian-zy0329/go-mall/logic/domainservice"
)

type UserAppSvc struct {
	ctx           context.Context
	userDomainSvc *domainservice.UserDomainSvc
}

func NewUserAppSvc(ctx context.Context) *UserAppSvc {
	return &UserAppSvc{
		ctx:           ctx,
		userDomainSvc: domainservice.NewUserDomainSvc(ctx),
	}
}

func (us *UserAppSvc) GenToken() (*reply.TokenReply, error) {
	token, err := us.userDomainSvc.GenAuthToken(12345678, "h5", "")
	if err != nil {
		return nil, err
	}
	logger.New(us.ctx).Info("generate token success", "tokenData", token)
	tokenReply := new(reply.TokenReply)
	util.CopyProperties(tokenReply, token)
	return tokenReply, err
}

func (us *UserAppSvc) TokenRefresh(refreshToken string) (*reply.TokenReply, error) {
	token, err := us.userDomainSvc.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	logger.New(us.ctx).Info("refresh token success", "tokenData", token)
	tokenReply := new(reply.TokenReply)
	util.CopyProperties(tokenReply, token)
	return tokenReply, err
}

func (us *UserAppSvc) UserRegister(userRegisterReq *request.UserRegister) error {
	userInfo := new(do.UserBaseInfo)
	util.CopyProperties(userInfo, userRegisterReq)
	_, err := us.userDomainSvc.RegisterUser(userInfo, userRegisterReq.Password)
	if errors.Is(err, errcode.ErrUserNameOccupied) {
		return err
	}
	if err != nil && !errors.Is(err, errcode.ErrUserNameOccupied) {
		return err
	}
	return nil
}

func (us *UserAppSvc) UserLogin(userLoginReq *request.UserLogin) (*reply.TokenReply, error) {
	token, err := us.userDomainSvc.LoginUser(userLoginReq.Body.LoginName, userLoginReq.Body.Password, userLoginReq.Header.Platform)
	if err != nil {
		return nil, err
	}
	logger.New(us.ctx).Info("login success", "tokenData", token)
	tokenReply := new(reply.TokenReply)
	util.CopyProperties(tokenReply, token)
	return tokenReply, nil
}

func (us *UserAppSvc) UserLogout(userId int64, platform string) error {
	err := us.userDomainSvc.LogoutUser(userId, platform)
	return err
}
