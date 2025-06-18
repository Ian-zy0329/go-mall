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

func (us *UserAppSvc) PasswordResetApply(request *request.PasswordResetApply) (*reply.PasswordResetApply, error) {
	passwordResetToken, code, err := us.userDomainSvc.ApplyForPasswordReset(request.LoginName)
	logger.New(us.ctx).Info("PasswordResetApply", "token", passwordResetToken, "code", code)
	if err != nil {
		return nil, err
	}
	reply := new(reply.PasswordResetApply)
	reply.PasswordResetToken = passwordResetToken
	return reply, nil
}

func (us *UserAppSvc) PasswordReset(request *request.PasswordReset) error {
	return us.userDomainSvc.PasswordReset(request.Token, request.Code, request.Password)
}

func (us *UserAppSvc) UserInfo(userId int64) *reply.UserInfoReply {
	userInfo := us.userDomainSvc.GetUserBaseInfo(userId)
	if userInfo == nil || userInfo.ID == 0 {
		return nil
	}
	infoReply := new(reply.UserInfoReply)
	util.CopyProperties(infoReply, userInfo)
	infoReply.LoginName = util.MaskLoginName(userInfo.LoginName)
	return infoReply
}

func (us *UserAppSvc) UserInfoUpdate(request *request.UserInfoUpdate, userId int64) error {
	return us.userDomainSvc.UpdateUserBaseInfo(request, userId)
}

func (us *UserAppSvc) AddUserAddress(request *request.UserAddress, userId int64) error {
	userAddressInfo := new(do.UserAddressInfo)
	err := util.CopyProperties(userAddressInfo, request)
	if err != nil {
		return errcode.Wrap("请求转换成领域对象失败", err)
	}
	userAddressInfo.UserId = userId
	newUserAddress, err := us.userDomainSvc.AddUserAddress(userAddressInfo)
	if err != nil {
		logger.New(us.ctx).Error("添加用户收获地址失败", "err", err, "return data", newUserAddress)
	}
	return err
}

func (us *UserAppSvc) GetSingleAddress(addressId int64) (*reply.UserAddress, error) {
	userAddress := new(reply.UserAddress)
	address, err := us.userDomainSvc.GetSingleAddress(addressId)
	if err != nil {
		return nil, errcode.Wrap("获取用户收获地址失败", err)
	}
	err = util.CopyProperties(&userAddress, address)
	if err != nil {
		errcode.Wrap("GetSingleAddress转换响应数据时发生错误", err)
		return nil, err
	}
	userAddress.MaskUserName = util.MaskRealName(address.UserName)
	userAddress.MaskUserPhone = util.MaskRealName(address.UserPhone)
	return userAddress, nil
}

func (us *UserAppSvc) GetUserAddresses(userId int64) ([]*reply.UserAddress, error) {
	userAddresses := make([]*reply.UserAddress, 0)
	addresses, err := us.userDomainSvc.GetUserAddresses(userId)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return userAddresses, nil
	}
	err = util.CopyProperties(&userAddresses, addresses)
	if err != nil {
		errcode.Wrap("GetUserAddressesError转换响应数据时发生错误", err)
		return nil, err
	}
	for _, address := range userAddresses {
		address.MaskUserName = util.MaskRealName(address.UserName)
		address.MaskUserPhone = util.MaskPhone(address.UserPhone)
	}
	return userAddresses, nil
}

func (us *UserAppSvc) ModifyUserAddress(request *request.UserAddress, userId int64, addressId int64) error {
	userAddressInfo := new(do.UserAddressInfo)
	err := util.CopyProperties(&userAddressInfo, request)
	if err != nil {
		err = errcode.Wrap("转换错误", err)
		return err
	}
	userAddressInfo.UserId = userId
	userAddressInfo.ID = addressId
	err = us.userDomainSvc.ModifyUserAddress(userAddressInfo)
	return err
}

func (us *UserAppSvc) DeleteUserAddress(userId, addressId int64) error {
	return us.userDomainSvc.DeleteUserAddress(userId, addressId)
}
