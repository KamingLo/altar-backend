package services

import (
	"altar/config"
	"altar/models"
	"altar/utils"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func CekAsdos(userID string) *string {
	var asdos models.AsistenDosen
	if err := config.DB.Where("user_id = ?", userID).First(&asdos).Error; err != nil {
		return nil
	}
	return &asdos.ID
}

func CekKoordinator(userID string) *string {
	var koor models.Koordinator
	if err := config.DB.Where("user_id = ?", userID).First(&koor).Error; err != nil {
		return nil
	}
	return &koor.ID
}

func LoginUser(input models.UserLogin) (string, error) {
	var user models.User

	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return "", errors.New("email not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return "", errors.New("incorrect password")
	}

	idAsisten := CekAsdos(user.ID)
	idKoordinator := CekKoordinator(user.ID)

	return utils.GenerateToken(user.ID, user.Email, idAsisten, idKoordinator)
}

func HandleGoogleLogin(email string) (string, error) {
	var user models.User

	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("user not found")
	}

	idAsisten := CekAsdos(user.ID)
	idKoordinator := CekKoordinator(user.ID)

	return utils.GenerateToken(user.ID, user.Email, idAsisten, idKoordinator)
}

func ForgotPassword(email string) error {
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("email tidak ditemukan")
	}

	var resetRecord models.PasswordReset
	errDB := config.DB.Where("email = ?", email).First(&resetRecord).Error

	now := time.Now()
	// Generate token unik di sini
	newToken := utils.GenerateCustomID("tok", 12)

	if errDB == nil {
		// Logika Cooldown (Rate Limiting)
		var cooldown time.Duration
		switch resetRecord.RequestCount {
		case 1:
			cooldown = 30 * time.Second
		case 2:
			cooldown = 60 * time.Second
		case 3:
			cooldown = 5 * time.Minute
		default:
			cooldown = 1 * time.Hour
		}

		if time.Since(resetRecord.UpdatedAt) < cooldown {
			return fmt.Errorf("terlalu banyak permintaan; coba lagi nanti")
		}

		// Update record: ganti token lama dengan yang baru
		resetRecord.Token = newToken
		resetRecord.ExpiredAt = now.Add(15 * time.Minute)
		resetRecord.RequestCount += 1
		resetRecord.UpdatedAt = now

		if err := config.DB.Save(&resetRecord).Error; err != nil {
			return errors.New("gagal memperbarui sesi reset")
		}
	} else {
		// Buat record baru
		newReset := models.PasswordReset{
			Email:        email,
			Token:        newToken,
			ExpiredAt:    now.Add(15 * time.Minute),
			RequestCount: 1,
			UpdatedAt:    now,
		}
		if err := config.DB.Create(&newReset).Error; err != nil {
			return errors.New("gagal membuat sesi reset")
		}
		resetRecord = newReset
	}

	// Kirim Link dengan Token rahasia ke Email
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s&email=%s",
		os.Getenv("FRONTEND_URL"), resetRecord.Token, email)

	return SendForgotPasswordLink(email, resetURL)
}

func ResetPassword(email, token, newPassword string) error {
	var resetRecord models.PasswordReset

	// 1. Verifikasi Email DAN Token (Kunci Keamanan Utama)
	if err := config.DB.Where("email = ? AND token = ?", email, token).First(&resetRecord).Error; err != nil {
		return errors.New("tautan tidak valid atau sudah kedaluwarsa")
	}

	// 2. Cek Kedaluwarsa Waktu
	if time.Now().After(resetRecord.ExpiredAt) {
		return errors.New("tautan sudah kedaluwarsa")
	}

	// 3. Update Password User
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("user tidak ditemukan")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	if err := config.DB.Save(&user).Error; err != nil {
		return errors.New("gagal menyimpan password baru")
	}

	// 4. Hapus record reset (Sifatnya sekali pakai)
	config.DB.Delete(&resetRecord)

	return nil
}
