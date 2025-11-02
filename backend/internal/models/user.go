package models

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/repository"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	RoleID   uint   `gorm:"not null,default:1" json:"role_id"`
	Role     Role   `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
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

func (u *User) FindByID(rep repository.Repository, id int) (*User, error) {
	if err := rep.Where("id = ?", id).First(u).Error; err != nil {
		return nil, appErrors.ErrUserNotFound
	}
	return u, nil
}

func (u *User) Create(rep repository.Repository) error {
	if err := rep.Create(u).Error; err != nil {
		return appErrors.ErrInvalidServer
	}
	return nil
}

func (u *User) Update(rep repository.Repository, upd map[string]interface{}) error {
	if err := rep.Model(u).Updates(upd).Error; err != nil {
		return appErrors.ErrUpdDataUser
	}
	return nil
}
