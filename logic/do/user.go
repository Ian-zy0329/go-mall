package do

import "time"

type SessionInfo struct {
	UserId       int64  `json:"user_id"`
	Platform     string `json:"platform"`
	SessionId    string `json:"session_id"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenInfo struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	Duration      int64     `json:"duration"`
	SrvCreateTime time.Time `json:"srv_create_time"`
}

type UserBaseInfo struct {
	ID        uint64    `json:"id"`
	NickName  string    `json:"nick_name"`
	LoginName string    `json:"login_name"`
	Verified  string    `json:"verified"`
	Avatar    string    `json:"avatar"`
	Slogan    string    `json:"slogan"`
	IsBlocked uint      `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenVerify struct {
	Approved  bool   // 验证结果
	UserId    int64  // 用户ID
	Platform  string // 用户Token对应的登录平台
	SessionId string // SessionId 可以用于存储一些与登录相关的东西, 用户不重新登录不会变
}

type UserAddressInfo struct {
	ID            int64
	UserId        int64
	UserName      string
	UserPhone     string
	Default       int
	ProvinceName  string
	CityName      string
	RegionName    string
	DetailAddress string
	IsDel         int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
