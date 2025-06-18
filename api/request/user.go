package request

type UserRegister struct {
	LoginName       string `json:"login_name"binding:"required,e164|email"`
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" binding:"required,eqfield=Password"`
	Nickname        string `json:"nickname"binding:"max=30"`
	Slogan          string `json:"slogan"binding:"max=100"`
	Avatar          string `json:"avatar"binding:"max=100"`
}

type UserLogin struct {
	Body struct {
		LoginName string `json:"login_name"binding:"required,e164|email"`
		Password  string `json:"password"binding:"required,min=8"`
	}
	Header struct {
		Platform string `json:"platform" header:"platform"binding:"required,oneof=H5 APP"`
	}
}

type PasswordResetApply struct {
	LoginName string `json:"login_name"binding:"required,e164|email"`
}

type PasswordReset struct {
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" binding:"required,eqfield=Password"`
	Token           string `json:"password_reset_token" binding:"required"`
	Code            string `json:"password_reset_code" binding:"required"`
}

type UserInfoUpdate struct {
	Nickname string `json:"nickname"binding:"max=30"`
	Slogan   string `json:"slogan"binding:"max=100"`
	Avatar   string `json:"avatar"binding:"max=100"`
}

type UserAddress struct {
	UserName      string `json:"user_name"binding:"required"`
	UserPhone     string `json:"user_phone"binding:"required"`
	Default       int    `json:"default"binding:"oneof=0 1"`
	ProvinceName  string `json:"province_name"binding:"required"`
	CityName      string `json:"city_name"binding:"required"`
	RegionName    string `json:"region_name"binding:"required"`
	DetailAddress string `json:"detail_address"binding:"required"`
}
