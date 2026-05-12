package dto

type TelegramAuthRequest struct {
	InitData         string `json:"init_data"`
	TelegramID       int64  `json:"telegram_id"`
	TelegramUsername string `json:"telegram_username"`
	DisplayName      string `json:"display_name"`
	AvatarURL        string `json:"avatar_url"`
	AuthDate         int64  `json:"auth_date"`
	Hash             string `json:"hash"`
}
