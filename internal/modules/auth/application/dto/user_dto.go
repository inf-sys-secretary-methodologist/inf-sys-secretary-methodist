package dto

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=12,max=128"`
	Role     string `json:"role" binding:"required,oneof=admin secretary methodist teacher student"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=1,max=128"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
