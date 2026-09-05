package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model          // Already includes "id", "created_at", "updated_at", "deleted_at"
	Username     string `json:"username" bson:"username"`
	Email        string `json:"email" bson:"email"`
	PasswordHash string `json:"password_hash" bson:"password_hash"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Saves a user to the database
func (user *User) Save() (*User, error) {
	err := Database.Model(&user).Create(&user).Error
	if err != nil {
		return &User{}, err
	}
	return user, nil
}

// Fetches a user by email from the database
func FetchUserByEmail(email string) (*User, error) {
	var user User
	err := Database.Model(&User{}).Where("email = ?", email).First(&user).Error
	if err != nil {
		return &User{}, err
	}
	return &user, nil
}

// Checks if the provided password matches the hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Updates a user in the database
func (user *User) UpdateUser(id string) (*User, error) {
	err := Database.Model(&User{}).Where("id = ?", id).Updates(user).Error
	if err != nil {
		return &User{}, err
	}
	return user, nil
}

// Deletes a user from the database
func DeleteUser(id string) error {
	err := Database.Model(&User{}).Where("id = ?", id).Delete(&User{}).Error
	if err != nil {
		return err
	}
	return nil
}
