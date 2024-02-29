package repository

import (
	"frontend-main/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) CreateUser(user *models.User) {
	ur.DB.Create(user)
}

func (ur *UserRepository) GetUserByID(id uint) *models.User {
	var user models.User
	ur.DB.First(&user, id)
	return &user
}

func (ur *UserRepository) UpdateUserNameByID(id uint, newName string) {
	var user models.User
	ur.DB.First(&user, id)
	ur.DB.Model(&user).Update("Name", newName)
}

func (ur *UserRepository) DeleteUserByID(id uint) {
	var user models.User
	ur.DB.Delete(&user, id)
}

func (ur *UserRepository) GetAllUsers() []models.User {
	var users []models.User
	ur.DB.Find(&users)
	return users
}
