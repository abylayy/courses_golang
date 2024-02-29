package repository

import (
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) CreateUser(user *User) {
	ur.DB.Create(user)
}

func (ur *UserRepository) GetUserByID(id uint) *User {
	var user User
	ur.DB.First(&user, id)
	return &user
}

func (ur *UserRepository) UpdateUserNameByID(id uint, newName string) {
	var user User
	ur.DB.First(&user, id)
	ur.DB.Model(&user).Update("Name", newName)
}

func (ur *UserRepository) DeleteUserByID(id uint) {
	var user User
	ur.DB.Delete(&user, id)
}

func (ur *UserRepository) GetAllUsers() []User {
	var users []User
	ur.DB.Find(&users)
	return users
}
