package dto

type RegisterInput struct {
	Name     string `json:"name" binding:"required,min=2,max=255" validate:"no_xss"`
	Email    string `json:"email" binding:"required,email,max=255" validate:"no_xss,no_sql_injection"`
	Password string `json:"password" binding:"required,min=8,max=128" validate:"strong_password"`
	Role     string `json:"role" binding:"required,oneof=system_admin methodist academic_secretary teacher student" validate:"no_xss,no_sql_injection"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255" validate:"no_xss,no_sql_injection"`
	Password string `json:"password" binding:"required,min=1,max=128"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required" validate:"no_xss,no_sql_injection"`
}
