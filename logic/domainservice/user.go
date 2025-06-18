package domainservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/cache"
	"github.com/Ian-zy0329/go-mall/dal/dao"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"time"
)

type UserDomainSvc struct {
	ctx     context.Context
	userDao *dao.UserDao
}

func NewUserDomainSvc(ctx context.Context) *UserDomainSvc {
	return &UserDomainSvc{
		ctx:     ctx,
		userDao: dao.NewUserDao(ctx),
	}
}

func (us *UserDomainSvc) RegisterUser(userInfo *do.UserBaseInfo, plainPassword string) (*do.UserBaseInfo, error) {
	existedUser, err := us.userDao.FindUserByLoginName(userInfo.LoginName)
	if err != nil {
		return nil, errcode.Wrap("UserDomainSvcRegisterUserError", err)
	}
	if existedUser.LoginName != "" {
		return nil, errcode.ErrUserNameOccupied
	}
	passwordHash, err := util.BcryptPassword(plainPassword)
	if err != nil {
		err = errcode.Wrap("UserDomainSvcRegisterUserError", err)
		return nil, err
	}
	userModel, err := us.userDao.CreateUser(userInfo, passwordHash)
	if err != nil {
		err = errcode.Wrap("UserDomainSvcRegisterUserError", err)
		return nil, err
	}
	err = util.CopyProperties(userInfo, userModel)
	if err != nil {
		err = errcode.Wrap("UsserDomainSvcRegisterUSerError", err)
		return nil, err
	}
	return userInfo, nil
}

func (us *UserDomainSvc) GetUserBaseInfo(userId int64) *do.UserBaseInfo {
	user, err := us.userDao.FindUserById(userId)
	log := logger.New(us.ctx)
	if err != nil {
		log.Error("GetUserBaseInfoError", "err", err)
		return nil
	}
	userBaseInfo := new(do.UserBaseInfo)
	err = util.CopyProperties(userBaseInfo, user)
	if err != nil {
		log.Error("GetUserBaseInfoError", "err", err)
		return nil
	}
	return userBaseInfo
}

func (us *UserDomainSvc) GenAuthToken(userId int64, platform string, sessionId string) (*do.TokenInfo, error) {
	user := us.GetUserBaseInfo(userId)
	if user.ID == 0 || user.IsBlocked == enum.UserBlockStateBlocked {
		err := errcode.ErrUserInvalid
		return nil, err
	}
	userSession := new(do.SessionInfo)
	userSession.UserId = userId
	userSession.Platform = platform
	if sessionId != "" {
		sessionId = util.GenSessionId(userId)
	}
	userSession.SessionId = sessionId
	accessToken, refreshToken, err := util.GenUserAuthToken(userId)
	userSession.AccessToken = accessToken
	userSession.RefreshToken = refreshToken
	if err != nil {
		err = errcode.Wrap("生成Token失败", err)
		return nil, err
	}
	err = cache.SetUserToken(us.ctx, userSession)
	if err != nil {
		err = errcode.Wrap("缓存Token失败", err)
		return nil, err
	}
	err = cache.DelOldSessionToken(us.ctx, userSession)
	if err != nil {
		errcode.Wrap("删除旧Token失败", err)
		return nil, err
	}
	err = cache.SetUserSession(us.ctx, userSession)
	if err != nil {
		err = errcode.Wrap("缓存Session失败", err)
		return nil, err
	}
	srvCreateTime := time.Now()
	tokenInfo := &do.TokenInfo{
		AccessToken:   userSession.AccessToken,
		RefreshToken:  userSession.RefreshToken,
		Duration:      int64(enum.AccessTokenDuration.Seconds()),
		SrvCreateTime: srvCreateTime,
	}
	return tokenInfo, nil
}

func (us *UserDomainSvc) VerifyAuthToken(accessToken string) (*do.TokenVerify, error) {
	tokenInfo, err := cache.GetAccessToken(us.ctx, accessToken)
	if err != nil {
		logger.New(us.ctx).Error("GetAccessTokenErr", "err", err)
		return nil, err
	}
	tokenVerify := new(do.TokenVerify)
	if tokenInfo != nil && tokenInfo.UserId != 0 {
		tokenVerify.Approved = true
		tokenVerify.UserId = tokenInfo.UserId
		tokenVerify.Platform = tokenInfo.Platform
		tokenVerify.SessionId = tokenInfo.SessionId
	} else {
		tokenVerify.Approved = false
	}
	return tokenVerify, nil
}

func (us *UserDomainSvc) RefreshToken(refreshToken string) (*do.TokenInfo, error) {
	log := logger.New(us.ctx)
	ok, err := cache.LockTokenRefresh(us.ctx, refreshToken)
	defer cache.UnlockTokenRefresh(us.ctx, refreshToken)
	if err != nil {
		err = errcode.Wrap("刷新Token时设置Redis锁发生错误", err)
		return nil, err
	}
	if !ok {
		err = errcode.ErrTooManyRequests
		return nil, err
	}
	tokenSessin, err := cache.GetRefreshToken(us.ctx, refreshToken)
	if err != nil {
		log.Error("GetRefreshTokenCacheErr", "err", err)
		err = errcode.ErrToken
		return nil, err
	}
	if tokenSessin == nil || tokenSessin.UserId == 0 {
		err = errcode.ErrToken
		return nil, err
	}
	userSession, err := cache.GetUserPlatformSession(us.ctx, tokenSessin.UserId, tokenSessin.Platform)
	if err != nil {
		log.Error("GetUserPlatformSessionCacheErr", "err", err)
		err = errcode.ErrToken
		return nil, err
	}
	if userSession.RefreshToken != refreshToken {
		log.Warn("ExpiredRefreshToken", "requestToken", refreshToken, "newToken", userSession.RefreshToken, "userId", userSession.UserId)
		err = errcode.ErrToken
		return nil, err
	}
	tokenInfo, err := us.GenAuthToken(tokenSessin.UserId, tokenSessin.Platform, tokenSessin.SessionId)
	if err != nil {
		err = errcode.Wrap("GenAuthTokenErr", err)
		return nil, err
	}
	return tokenInfo, nil
}

func (us *UserDomainSvc) LoginUser(LoginName, plainPassword, platform string) (*do.TokenInfo, error) {
	existedUser, err := us.userDao.FindUserByLoginName(LoginName)
	if err != nil {
		return nil, errcode.Wrap("UserDomainSvcLoginUserError", err)
	}
	if existedUser.ID == 0 {
		return nil, errcode.ErrUserNotRight
	}
	if !util.BcryptCompare(existedUser.Password, plainPassword) {
		return nil, errcode.ErrUserNotRight
	}

	tokenInfo, err := us.GenAuthToken(existedUser.ID, platform, "")
	return tokenInfo, err
}

func (us *UserDomainSvc) LogoutUser(userId int64, platform string) error {
	log := logger.New(us.ctx)
	userSession, err := cache.GetUserPlatformSession(us.ctx, userId, platform)
	if err != nil {
		log.Error("GetUserPlatformSessionCacheErr", "err", err)
		return errcode.Wrap("UserDomainSvcLogoutUserError", err)
	}

	err = cache.DelAccessToken(us.ctx, userSession.AccessToken)
	if err != nil {
		log.Error("DelAccessTokenCacheErr", "err", err)
		return errcode.Wrap("UserDomainSvcLogoutUserError", err)
	}
	err = cache.DelRefreshToken(us.ctx, userSession.RefreshToken)
	if err != nil {
		log.Error("DelRefreshTokenCacheErr", "err", err)
		return errcode.Wrap("UserDomainSvcLogoutUserError", err)
	}

	err = cache.DelUserSessionOnPlatform(us.ctx, userSession)
	if err != nil {
		log.Error("DelUserSessionCacheErr", "err", err)
		return errcode.Wrap("UserDomainSvcLogoutUserError", err)
	}
	return nil
}

func (us *UserDomainSvc) ApplyForPasswordReset(loginName string) (passwordResetToken, code string, err error) {
	user, err := us.userDao.FindUserByLoginName(loginName)
	if err != nil {
		err = errcode.Wrap("ApplyForPasswordResetError", err)
		return
	}
	if user.ID == 0 {
		err = errcode.ErrUserNotRight
		return
	}
	token, err := util.GenPasswordResetToken(user.ID)
	code = util.RandNumStr(6)
	if err != nil {
		err = errcode.Wrap("ApplyForPasswordResetError", err)
		return
	}
	err = cache.SetPasswordRessetToken(us.ctx, user.ID, token, code)
	if err != nil {
		err = errcode.Wrap("ApplyForPasswordResetError", err)
		return
	}
	passwordResetToken = token
	return
}

func (us *UserDomainSvc) PasswordReset(token, code, newPassword string) error {
	log := logger.New(us.ctx)
	userId, resetCode, err := cache.GetPasswordResetToken(us.ctx, token)
	if err != nil {
		log.Error("ResetPasswordError", "err", err)
		err = errcode.Wrap("ResetPasswordError", err)
		return err
	}
	if userId == 0 || resetCode != code {
		err = errcode.ErrParams
		return err
	}
	user, err := us.userDao.FindUserById(userId)
	if err != nil {
		return errcode.Wrap("ResetPasswordError", err)
	}
	if user.ID == 0 || user.IsBlocked == enum.UserBlockStateBlocked {
		err = errcode.ErrUserInvalid
		return err
	}
	newPass, err := util.BcryptPassword(newPassword)
	if err != nil {
		return errcode.Wrap("ResetPasswordError", err)
	}
	user.Password = newPass
	err = us.userDao.UpdateUser(user)
	if err != nil {
		return errcode.Wrap("ResetPasswordError", err)
	}
	err = cache.DelUserSessions(us.ctx, user.ID)
	if err != nil {
		log.Error("ResetPasswordError", "err", err)
	}
	err = cache.DelPasswordResetToken(us.ctx, token)
	if err != nil {
		log.Error("ResetPasswordError", "err", err)
	}
	return nil
}

func (us *UserDomainSvc) UpdateUserBaseInfo(request *request.UserInfoUpdate, userId int64) error {
	user, err := us.userDao.FindUserById(userId)
	if err != nil {
		return err
	}
	user.Avatar = request.Avatar
	user.Nickname = request.Nickname
	user.Slogan = request.Slogan
	err = us.userDao.UpdateUser(user)
	return err
}

func (us *UserDomainSvc) AddUserAddress(addressInfo *do.UserAddressInfo) (*do.UserAddressInfo, error) {
	addressModel, err := us.userDao.CreateUserAddress(addressInfo)
	if err != nil {
		err = errcode.Wrap("UserDomainSvcAddUserAddressError", err)
		return nil, err
	}
	err = util.CopyProperties(addressInfo, addressModel)
	if err != nil {
		err = errcode.Wrap("AddUserAddressError", err)
		return nil, err
	}
	return addressInfo, nil
}

func (us *UserDomainSvc) GetSingleAddress(addressId int64) (*do.UserAddressInfo, error) {
	addressModel, err := us.userDao.GetSingleAddress(addressId)
	if err != nil {
		err = errcode.Wrap("UserDomainSvcGetSingleAddressError", err)
		return nil, err
	}
	userAddress := new(do.UserAddressInfo)
	err = util.CopyProperties(&userAddress, addressModel)
	if err != nil {
		err = errcode.Wrap("UserDomainSvcGetSingleAddressError", err)
		return nil, err
	}
	return userAddress, nil
}

func (us *UserDomainSvc) GetUserAddresses(userId int64) ([]*do.UserAddressInfo, error) {
	userAddresses, err := us.userDao.GetUserAddresses(userId)
	if err != nil {
		err = errcode.Wrap("GetUserAddressesError", err)
		return nil, err
	}
	addresses := make([]*do.UserAddressInfo, 0)
	err = util.CopyProperties(&addresses, userAddresses)
	if err != nil {
		err = errcode.Wrap("GetUserAddressesError转换响应数据时fa", err)
		return nil, err
	}
	return addresses, nil
}

func (us *UserDomainSvc) ModifyUserAddress(address *do.UserAddressInfo) error {
	addressModel, err := us.userDao.GetSingleAddress(address.ID)
	log := logger.New(us.ctx)
	if err != nil || address.UserId != addressModel.UserId {
		log.Error("UserAddressNotMatchError", "err", err, "return data", address, "request data", address)
		return errcode.ErrParams
	}
	err = us.userDao.UpdateUserAddress(address)
	if err != nil {
		err = errcode.Wrap("UpdateUserAddressError", err)
	}
	return err
}

func (us *UserDomainSvc) DeleteUserAddress(userId, addressId int64) error {
	address, err := us.userDao.GetSingleAddress(addressId)
	if err != nil || address.UserId != userId {
		logger.New(us.ctx).Error("UserAddressNotMatchError", "err", err, "return data", address, "addressId", addressId, "userId", userId)
		return errcode.ErrParams
	}
	err = us.userDao.DeleteOneAddress(address)
	return err
}
