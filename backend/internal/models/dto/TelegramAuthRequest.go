package dto

type TelegramAuthRequest struct {
	TelegramID       int64  `json:"telegram_id" binding:"required"`
	TelegramUsername string `json:"telegram_username"`
	DisplayName      string `json:"display_name" binding:"required"`
	AvatarURL        string `json:"avatar_url"`
	AuthDate         int64  `json:"auth_date" binding:"required"`
	Hash             string `json:"hash" binding:"required"`
}
