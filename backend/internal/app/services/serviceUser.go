package services

import (
	"cinema/internal/app/utils"
	"cinema/internal/container"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
)

type UserServices interface {
}

type userServices struct {
	container container.Container
}

func NewUserService(container container.Container) UserServices {
	return &userServices{
		container: container,
	}
}

func (u *userServices) UpdateInf(oldName, username, password string) (tokenPair dto.TokenPair, err error) {
	user, err := new(models.User).FindByName(u.container.GetRepository(), oldName)
	if err != nil {
		return tokenPair, err
	}
	updates := make(map[string]interface{})

	if username != "" {
		updates["username"] = username
	}

	if password != "" {
		hashed, _ := utils.HashPassword(password)
		updates["password"] = hashed
	}

	if len(updates) == 0 {
		return tokenPair, appErrors.ErrNotEnougthData
	}

	err = user.Update(u.container.GetRepository(), updates)
	if err != nil {
		return tokenPair, err
	}
	tokenPair, err = createTokens(u.container.GetLogger(), *u.container.GetConfig(), user)
	return tokenPair, nil
}
