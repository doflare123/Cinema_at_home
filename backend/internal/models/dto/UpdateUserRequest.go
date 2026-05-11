package dto

type UpdateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password" binding:"omitempty,password"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}
