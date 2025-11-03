package dto

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email,max=255,no_xss,no_sql_injection"`
	Password string `json:"password" binding:"required,strong_password,max=128"`
	Role     string `json:"role" binding:"required,oneof=admin secretary methodist teacher student"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255,no_xss,no_sql_injection"`
	Password string `json:"password" binding:"required,min=1,max=128"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required,no_xss,no_sql_injection"`
}
