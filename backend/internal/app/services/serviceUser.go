package services

import (
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
)

type UserServices interface {
	UpdateInf(userID uint, username, password, displayName, avatarURL string) (dto.TokenPair, error)
}

type userServices struct {
	container container.Container
}

func NewUserService(container container.Container) UserServices {
	return &userServices{container: container}
}

func (u *userServices) UpdateInf(userID uint, username, password, displayName, avatarURL string) (tokenPair dto.TokenPair, err error) {
	user, err := new(models.User).FindByID(u.container.GetRepository(), int(userID))
	if err != nil {
		return tokenPair, err
	}
	updates := make(map[string]interface{})

	if username != "" {
		updates["username"] = username
	}
	if displayName != "" {
		updates["display_name"] = displayName
	}
	if avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}
	if password != "" {
		hashed, hashErr := utils.HashPassword(password)
		if hashErr != nil {
			return tokenPair, appErrors.ErrInvalidServer
		}
		updates["password"] = hashed
	}

	if len(updates) == 0 {
		return tokenPair, appErrors.ErrNotEnougthData
	}

	err = user.Update(u.container.GetRepository(), updates)
	if err != nil {
		if isUniqueViolation(err) {
			return tokenPair, appErrors.ErrUserNameAlreadyExist
		}
		return tokenPair, err
	}
	tokenPair, err = createTokens(u.container.GetLogger(), *u.container.GetConfig(), user)
	return tokenPair, err
}
