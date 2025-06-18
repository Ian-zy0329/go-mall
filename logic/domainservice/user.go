package domainservice

import (
	"context"
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
	return &do.UserBaseInfo{
		ID:        12345678,
		NickName:  "Ian",
		LoginName: "zy0329",
		Verified:  1,
		Avatar:    "",
		Slogan:    "",
		IsBlocked: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
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
