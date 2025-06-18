package reply

type TokenReply struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	Duration      int64  `json:"duration"`
	SrvCreateTime string `json:"srv_create_time"`
}

// PasswordResetApply 申请重置密码的响应
type PasswordResetApply struct {
	PasswordResetToken string `json:"password_reset_token"`
}

type UserInfoReply struct {
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	LoginName string `json:"login_name"`
	Verified  int    `json:"verified"`
	Avatar    string `json:"avatar"`
	Slogan    string `json:"slogan"`
	IsBlocked int    `json:"is_blocked"`
	CreatedAt string `json:"created_at"`
}

type UserAddress struct {
	ID            int64  `json:"id"`
	UserName      string `json:"user_name"`
	UserPhone     string `json:"user_phone"`
	MaskUserName  string `json:"mask_user_name"`
	MaskUserPhone string `json:"mask_user_phone"`
	Default       int    `json:"default"`
	ProvinceName  string `json:"province_name"`
	CityName      string `json:"city_name"`
	RegionName    string `json:"region_name"`
	DetailAddress string `json:"detail_address"`
	CreatedAt     string `json:"created_at"`
}
