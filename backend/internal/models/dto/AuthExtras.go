package dto

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserView struct {
	ID               uint   `json:"id"`
	Username         string `json:"username"`
	DisplayName      string `json:"display_name"`
	TelegramUsername string `json:"telegram_username,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	RoleID           uint   `json:"role_id"`
	Status           string `json:"status"`
}

type AdminUpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending active rejected blocked"`
}
