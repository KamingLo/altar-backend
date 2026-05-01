package main

import (
	"altar/config"
	"altar/models"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load()

	config.ConnectDatabase()

	email := "lokaming86@gmail.com"
	password := "admin123"
	username := "Kaming"
	nip := "535240175"
	nim := "535240175"

	tx := config.DB.Begin()

	var user models.User
	if err := tx.Where("email = ?", email).First(&user).Error; err == nil {
		fmt.Printf("User with email %s already exists.\n", email)
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to hash password: %v", err)
		}
		user = models.User{
			Username: username,
			Email:    email,
			Password: string(hashedPassword),
		}

		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			log.Fatalf("Failed to create user: %v", err)
		}
	}

	var existingKoordinator models.Koordinator
	if err := tx.Where("user_id = ?", user.ID).First(&existingKoordinator).Error; err == nil {
		fmt.Printf("Koordinator for user %s already exists. Skipping.\n", email)
	} else {
		koordinator := models.Koordinator{
			UserID: user.ID,
			NIP:    nip,
		}
		if err := tx.Create(&koordinator).Error; err != nil {
			tx.Rollback()
			log.Fatalf("Failed to create koordinator: %v", err)
		}
	}

	var existingAsdos models.AsistenDosen
	if err := tx.Where("user_id = ?", user.ID).First(&existingAsdos).Error; err == nil {
		fmt.Printf("Asisten Dosen for user %s already exists. Skipping.\n", email)
	} else {
		asistenDosen := models.AsistenDosen{
			UserID: user.ID,
			NIM:    nim,
		}
		if err := tx.Create(&asistenDosen).Error; err != nil {
			tx.Rollback()
			log.Fatalf("Failed to create asisten dosen: %v", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("==========================================")
	fmt.Println("Seeder Successful!")
	fmt.Println("User Email:      ", email)
	fmt.Println("User Password:   ", password)
	fmt.Println("==========================================")
	fmt.Println("Please change your password after logging in.")
}
