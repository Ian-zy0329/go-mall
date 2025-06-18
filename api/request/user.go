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
