package services

import (
	"altar/config"
	"altar/models"
	"altar/utils"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserSummary struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AsdosSummary struct {
	ID       string `json:"id_asdos" gorm:"column:id"`
	Username string `json:"username" gorm:"column:username"`
	NIM      string `json:"nim" gorm:"column:nim"`
}

type KoorSummary struct {
	ID       string `json:"id_koor" gorm:"column:id"`
	Username string `json:"username" gorm:"column:username"`
	NIP      string `json:"nip" gorm:"column:nip"`
}

// --- User CRUD ---

func CreateUser(input *models.User) error {
	input.Email = strings.ToLower(input.Email)

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return errors.New("user with this email already exists")
	}

	randomPassword := utils.GenerateRandomPassword(8)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	input.Password = string(hashedPassword)

	if err := config.DB.Create(input).Error; err != nil {
		return errors.New("failed to create user account")
	}

	SendNewAccountPassword(input.Email, randomPassword)
	return nil
}

func GetAllUsers(page int, search string) ([]UserSummary, error) {
	var users []UserSummary
	query := config.DB.Model(&models.User{})

	if search != "" {
		searchLower := strings.ToLower(search)
		query = query.Where("(username ILIKE ? OR email ILIKE ?)", "%"+search+"%", "%"+searchLower+"%")
	}

	offset := (page - 1) * 10
	if err := query.Limit(10).Offset(offset).Select("id, username, email").Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserByID(id string) (models.User, error) {
	var user models.User
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return user, err
	}
	user.Password = ""
	return user, nil
}

func UpdateUser(id string, username, email string) error {
	var user models.User
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if new email is already taken by another user
	email = strings.ToLower(email)
	var existingUser models.User
	if err := config.DB.Where("email = ? AND id != ?", email, id).First(&existingUser).Error; err == nil {
		return errors.New("email already in use by another account")
	}

	user.Username = username
	user.Email = email

	return config.DB.Save(&user).Error
}

func DeleteUser(id string) error {
	return config.DB.Unscoped().Delete(&models.User{}, "id = ?", id).Error
}

// --- Asisten Dosen CRUD ---

func CreateAsdos(userID, nim, phone string) error {
	tx := config.DB.Begin()

	// Check if user exists
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if already an Asdos
	var existingAsdos models.AsistenDosen
	if err := tx.Where("user_id = ?", userID).First(&existingAsdos).Error; err == nil {
		tx.Rollback()
		return errors.New("user is already registered as Asisten Dosen")
	}

	asdos := models.AsistenDosen{
		UserID:      userID,
		NIM:         nim,
		PhoneNumber: phone,
	}

	if err := tx.Create(&asdos).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to create asisten dosen detail")
	}

	return tx.Commit().Error
}

func GetAllAsdos(page int, search string) ([]AsdosSummary, error) {
	var asdos []AsdosSummary
	query := config.DB.Model(&models.AsistenDosen{}).
		Select("asisten_dosens.id, users.username, asisten_dosens.nim").
		Joins("left join users on users.id = asisten_dosens.user_id").
		Where("users.deleted_at IS NULL")

	if search != "" {
		query = query.Where("(users.username ILIKE ? OR asisten_dosens.nim LIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	offset := (page - 1) * 10
	if err := query.Limit(10).Offset(offset).Scan(&asdos).Error; err != nil {
		return nil, err
	}
	return asdos, nil
}

func GetAsdosByID(id string) (models.AsistenDosen, error) {
	var asdos models.AsistenDosen
	if err := config.DB.Preload("User").Where("id = ?", id).First(&asdos).Error; err != nil {
		return asdos, err
	}
	return asdos, nil
}

func UpdateAsdos(id string, nim, phone string) error {
	var asdos models.AsistenDosen
	if err := config.DB.Where("id = ?", id).First(&asdos).Error; err != nil {
		return errors.New("asisten dosen not found")
	}

	asdos.NIM = nim
	asdos.PhoneNumber = phone

	return config.DB.Save(&asdos).Error
}

func DeleteAsdos(id string) error {
	tx := config.DB.Begin()

	var asdos models.AsistenDosen
	if err := tx.Where("id = ?", id).First(&asdos).Error; err != nil {
		tx.Rollback()
		return errors.New("asisten dosen not found")
	}

	userID := asdos.UserID

	if err := tx.Unscoped().Delete(&asdos).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if user has other roles
	var otherKoor models.Koordinator
	errKoor := tx.Where("user_id = ?", userID).First(&otherKoor).Error

	// If no other roles, delete user
	if errors.Is(errKoor, gorm.ErrRecordNotFound) {
		if err := tx.Unscoped().Delete(&models.User{}, "id = ?", userID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// --- Koordinator CRUD ---

func CreateKoordinator(userID, nip string) error {
	tx := config.DB.Begin()

	// Check if user exists
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if already a Koordinator
	var existingKoor models.Koordinator
	if err := tx.Where("user_id = ?", userID).First(&existingKoor).Error; err == nil {
		tx.Rollback()
		return errors.New("user is already registered as Koordinator")
	}

	koor := models.Koordinator{
		UserID: userID,
		NIP:    nip,
	}

	if err := tx.Create(&koor).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to create koordinator detail")
	}

	return tx.Commit().Error
}

func GetAllKoordinator(page int, search string) ([]KoorSummary, error) {
	var koor []KoorSummary
	query := config.DB.Model(&models.Koordinator{}).
		Select("koordinators.id, users.username, koordinators.nip").
		Joins("left join users on users.id = koordinators.user_id").
		Where("users.deleted_at IS NULL")

	if search != "" {
		query = query.Where("(users.username ILIKE ? OR koordinators.NIP LIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	offset := (page - 1) * 10
	if err := query.Limit(10).Offset(offset).Scan(&koor).Error; err != nil {
		return nil, err
	}
	return koor, nil
}

func GetKoordinatorByID(id string) (models.Koordinator, error) {
	var koor models.Koordinator
	if err := config.DB.Preload("User").Where("id = ?", id).First(&koor).Error; err != nil {
		return koor, err
	}
	return koor, nil
}

func UpdateKoordinator(id string, nip string) error {
	var koor models.Koordinator
	if err := config.DB.Where("id = ?", id).First(&koor).Error; err != nil {
		return errors.New("koordinator not found")
	}

	koor.NIP = nip

	return config.DB.Save(&koor).Error
}

func DeleteKoordinator(id string) error {
	tx := config.DB.Begin()

	var koor models.Koordinator
	if err := tx.Where("id = ?", id).First(&koor).Error; err != nil {
		tx.Rollback()
		return errors.New("koordinator not found")
	}

	userID := koor.UserID

	if err := tx.Unscoped().Delete(&koor).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if user has other roles
	var otherAsdos models.AsistenDosen
	errAsdos := tx.Where("user_id = ?", userID).First(&otherAsdos).Error

	// If no other roles, delete user
	if errors.Is(errAsdos, gorm.ErrRecordNotFound) {
		if err := tx.Unscoped().Delete(&models.User{}, "id = ?", userID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
