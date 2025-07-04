package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

func SetUserToken(ctx context.Context, session *do.SessionInfo) error {
	log := logger.New(ctx)
	err := setAccessToken(ctx, session)
	if err != nil {
		log.Error("redis error", "err", err)
		return err
	}
	err = setRefreshToken(ctx, session)
	if err != nil {
		log.Error("redis error", "err", err)
		return err
	}
	return err
}

func SetUserSession(ctx context.Context, session *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, session.UserId)
	sessionDataBytes, _ := json.Marshal(session)
	err := Redis().HSet(ctx, redisKey, session.Platform, sessionDataBytes).Err()
	if err != nil {
		logger.New(ctx).Error("redis error", "err", err)
		return err
	}
	return err
}

func DelOldSessionToken(ctx context.Context, session *do.SessionInfo) error {
	oldSession, err := GetUserPlatformSession(ctx, session.UserId, session.Platform)
	if err != nil {
		logger.New(ctx).Error("redis error", "err", err)
		return err
	}
	if oldSession == nil {
		return nil
	}
	err = DelAccessToken(ctx, oldSession.AccessToken)
	if err != nil {
		return errcode.Wrap("redis error", err)
	}
	err = DelayDelRefreshToken(ctx, oldSession.RefreshToken)
	if err != nil {
		return errcode.Wrap("redis error", err)
	}
	return nil
}

func GetUserPlatformSession(ctx context.Context, userId int64, platform string) (*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, userId)
	result, err := Redis().HGet(ctx, redisKey, platform).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	session := new(do.SessionInfo)
	err = json.Unmarshal([]byte(result), &session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func setAccessToken(ctx context.Context, session *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_ACCESS_TOKEN, session.AccessToken)
	sessionDataBytes, _ := json.Marshal(session)
	res, err := Redis().Set(ctx, redisKey, sessionDataBytes, enum.AccessTokenDuration).Result()
	logger.New(ctx).Debug("redis debug", "res", res, "err", err)
	return err
}

func GetAccessToken(ctx context.Context, accessToken string) (*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_ACCESS_TOKEN, accessToken)
	result, err := Redis().Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	session := new(do.SessionInfo)
	if errors.Is(err, redis.Nil) {
		return session, nil
	}
	json.Unmarshal([]byte(result), &session)
	return session, nil
}

func setRefreshToken(ctxx context.Context, session *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, session.RefreshToken)
	sessionDataBytes, _ := json.Marshal(session)
	return Redis().Set(ctxx, redisKey, sessionDataBytes, enum.RefreshTokenDuration).Err()
}

func DelAccessToken(ctx context.Context, accessToken string) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_ACCESS_TOKEN, accessToken)
	return Redis().Del(ctx, redisKey).Err()
}

func DelayDelRefreshToken(ctx context.Context, refreshToken string) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, refreshToken)
	return Redis().Expire(ctx, redisKey, enum.OldRefreshTokenHoldingDuration).Err()
}

func DelRefreshToken(ctx context.Context, refreshToken string) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, refreshToken)
	return Redis().Del(ctx, redisKey).Err()
}

func LockTokenRefresh(ctx context.Context, refreshToken string) (bool, error) {
	redisLockKey := fmt.Sprintf(enum.REDISKEY_TOKEN_REFRESH_LOCK, refreshToken)
	return Redis().SetNX(ctx, redisLockKey, "locked", 10*time.Second).Result()
}

func UnlockTokenRefresh(ctx context.Context, refreshToken string) error {
	redisLockKey := fmt.Sprintf(enum.REDISKEY_TOKEN_REFRESH_LOCK, refreshToken)
	return Redis().Del(ctx, redisLockKey).Err()
}

func GetRefreshToken(ctx context.Context, refreshToken string) (*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, refreshToken)
	result, err := Redis().Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	session := new(do.SessionInfo)
	if errors.Is(err, redis.Nil) {
		return session, nil
	}
	json.Unmarshal([]byte(result), &session)
	return session, nil
}

func DelUserSessionOnPlatform(ctx context.Context, userSession *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, userSession.UserId)
	return Redis().HDel(ctx, redisKey, userSession.Platform).Err()
}

func SetPasswordRessetToken(ctx context.Context, userId int64, token, code string) error {
	redisKey := fmt.Sprintf(enum.REDISKEY_PASSWORDRESET_TOKEN, token)
	val := fmt.Sprintf("%d:%s", userId, code)
	return Redis().Set(ctx, redisKey, val, enum.PasswordTokenDuration).Err()
}

func GetPasswordResetToken(ctx context.Context, token string) (userId int64, code string, err error) {
	redisKey := fmt.Sprintf(enum.REDISKEY_PASSWORDRESET_TOKEN, token)
	val, err := Redis().Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return
	}
	valArr := strings.Split(val, ":")
	userId, _ = strconv.ParseInt(valArr[0], 10, 64)
	code = valArr[1]
	return
}

func DelPasswordResetToken(ctx context.Context, token string) error {
	redisKey := fmt.Sprintf(enum.REDISKEY_PASSWORDRESET_TOKEN, token)
	return Redis().Del(ctx, redisKey).Err()
}

func DelUserSessions(ctx context.Context, userId int64) error {
	sessions, err := GetUserAllSessions(ctx, userId)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		DelOldSessionToken(ctx, session)
	}
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, userId)
	return Redis().Del(ctx, redisKey).Err()
}

func GetUserAllSessions(ctx context.Context, userId int64) (map[string]*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, userId)
	result, err := Redis().HGetAll(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	// key 不存在
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	sessions := make(map[string]*do.SessionInfo)
	for platform, sessionData := range result {
		session := new(do.SessionInfo)
		err = json.Unmarshal([]byte(sessionData), &session)
		if err != nil {
			return nil, err
		}
		sessions[platform] = session
	}
	//logger.New(ctx).Debug("hgetall user all session", "data", sessions)
	return sessions, nil
}
