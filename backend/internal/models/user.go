package models

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/repository"
)

type User struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	TelegramID       *int64 `gorm:"uniqueIndex" json:"telegram_id,omitempty"`
	TelegramUsername string `json:"telegram_username,omitempty"`
	DisplayName      string `gorm:"not null" json:"display_name"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	Username         string `gorm:"unique;not null" json:"username"`
	Password         string `gorm:"not null" json:"-"`
	RoleID           uint   `gorm:"not null;default:1" json:"role_id"`
	Status           string `gorm:"type:varchar(16);not null;default:'pending'" json:"status"`
	Role             Role   `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

func (u *User) NameAlreadyExist(rep repository.Repository, name string) error {
	if err := rep.Where("username = ?", name).First(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) FindByName(rep repository.Repository, name string) (*User, error) {
	if err := rep.Where("username = ?", name).First(u).Error; err != nil {
		return nil, appErrors.ErrUserNotFound
	}
	return u, nil
}

func (u *User) FindByTelegramID(rep repository.Repository, telegramID int64) (*User, error) {
	if err := rep.Where("telegram_id = ?", telegramID).First(u).Error; err != nil {
		return nil, appErrors.ErrUserNotFound
	}
	return u, nil
}

func (u *User) FindByID(rep repository.Repository, id int) (*User, error) {
	if err := rep.Where("id = ?", id).First(u).Error; err != nil {
		return nil, appErrors.ErrUserNotFound
	}
	return u, nil
}

func (u *User) Create(rep repository.Repository) error {
	if err := rep.Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) Update(rep repository.Repository, upd map[string]interface{}) error {
	if err := rep.Model(u).Updates(upd).Error; err != nil {
		return err
	}
	return nil
}

func (r *Role) FindByName(rep repository.Repository, name string) (*Role, error) {
	if err := rep.Where("name = ?", name).First(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}
