package models

type User struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Password uint64 `json:"password"`
	Salt     uint64 `json:"salt"`
	CreateAt string `json:"create_at"`
	UpdateAt string `json:"update_at"`
}
