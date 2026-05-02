package services

import (
	"altar/config"
	"altar/models"
	"altar/utils"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

// --- Asisten Dosen CRUD ---

func CreateAsdos(input *models.User, nim, phone string) error {
	tx := config.DB.Begin()

	var user models.User
	err := tx.Where("email = ?", input.Email).First(&user).Error

	if err == nil {
		// User exists, use existing ID
		input.ID = user.ID

		// Check if already an Asdos
		var existingAsdos models.AsistenDosen
		if err := tx.Where("user_id = ?", user.ID).First(&existingAsdos).Error; err == nil {
			tx.Rollback()
			return errors.New("user is already registered as Asisten Dosen")
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// User doesn't exist, create new
		randomPassword := utils.GenerateRandomPassword(8)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
		input.Password = string(hashedPassword)
		if err := tx.Create(input).Error; err != nil {
			tx.Rollback()
			return errors.New("failed to create user account")
		}
		SendNewAccountPassword(input.Email, randomPassword)
	} else {
		tx.Rollback()
		return err
	}

	asdos := models.AsistenDosen{
		UserID:      input.ID,
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
	query := config.DB.Table("asisten_dosens").
		Select("asisten_dosens.id, users.username, asisten_dosens.nim").
		Joins("left join users on users.id = asisten_dosens.user_id")

	if search != "" {
		query = query.Where("users.username ILIKE ? OR asisten_dosens.nim LIKE ?", "%"+search+"%", "%"+search+"%")
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

	if err := tx.Delete(&asdos).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if user has other roles
	var otherKoor models.Koordinator
	errKoor := tx.Where("user_id = ?", userID).First(&otherKoor).Error

	// If no other roles, delete user
	if errors.Is(errKoor, gorm.ErrRecordNotFound) {
		if err := tx.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// --- Koordinator CRUD ---

func CreateKoordinator(input *models.User, nip string) error {
	tx := config.DB.Begin()

	var user models.User
	err := tx.Where("email = ?", input.Email).First(&user).Error

	if err == nil {
		// User exists
		input.ID = user.ID

		// Check if already a Koordinator
		var existingKoor models.Koordinator
		if err := tx.Where("user_id = ?", user.ID).First(&existingKoor).Error; err == nil {
			tx.Rollback()
			return errors.New("user is already registered as Koordinator")
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// User doesn't exist
		randomPassword := utils.GenerateRandomPassword(8)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
		input.Password = string(hashedPassword)
		if err := tx.Create(input).Error; err != nil {
			tx.Rollback()
			return errors.New("failed to create user account")
		}
		SendNewAccountPassword(input.Email, randomPassword)
	} else {
		tx.Rollback()
		return err
	}

	koor := models.Koordinator{
		UserID: input.ID,
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
	query := config.DB.Table("koordinators").
		Select("koordinators.id, users.username, koordinators.nip").
		Joins("left join users on users.id = koordinators.user_id")

	if search != "" {
		query = query.Where("users.username ILIKE ? OR koordinators.NIP LIKE ?", "%"+search+"%", "%"+search+"%")
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

	if err := tx.Delete(&koor).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if user has other roles
	var otherAsdos models.AsistenDosen
	errAsdos := tx.Where("user_id = ?", userID).First(&otherAsdos).Error

	// If no other roles, delete user
	if errors.Is(errAsdos, gorm.ErrRecordNotFound) {
		if err := tx.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
