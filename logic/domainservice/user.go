package domainservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/cache"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"time"
)

type UserDomainSvc struct {
	ctx context.Context
}

func NewUserDomainSvc(ctx context.Context) *UserDomainSvc {
	return &UserDomainSvc{
		ctx: ctx,
	}
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
