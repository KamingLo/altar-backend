package controllers

import (
	"altar/config"
	"altar/models"
	"altar/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type KioskPINRequest struct {
	PIN string `json:"pin" binding:"required,min=4,max=6"`
}

func SetKioskPIN(c *gin.Context) {
	idKoor := c.GetString("id_koordinator")
	var req KioskPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid PIN format (4-6 digits)", err)
		return
	}

	hashedPIN, _ := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	hashedStr := string(hashedPIN)

	if err := config.DB.Model(&models.Koordinator{}).Where("id = ?", idKoor).Update("kiosk_pin", hashedStr).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to update PIN", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Kiosk PIN updated successfully", nil)
}

func VerifyKioskPIN(c *gin.Context) {
	idKoor := c.GetString("id_koordinator")
	var req KioskPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid PIN format", err)
		return
	}

	var koor models.Koordinator
	if err := config.DB.Where("id = ?", idKoor).First(&koor).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Koordinator not found", nil)
		return
	}

	if koor.KioskPIN == nil {
		utils.SendError(c, http.StatusForbidden, "Kiosk PIN not set", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*koor.KioskPIN), []byte(req.PIN)); err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Incorrect PIN", nil)
		return
	}

	// Re-issue token with is_kiosk_mode = true
	var user models.User
	config.DB.Where("id = ?", koor.UserID).First(&user)

	idAsisten := c.GetString("id_asisten")
	var asdosID *string
	if idAsisten != "" {
		asdosID = &idAsisten
	}

	token, err := utils.GenerateToken(user.ID, user.Email, asdosID, &koor.ID, true)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to generate kiosk token", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Kiosk Mode Activated", gin.H{"token": token})
}

func GenerateQR(c *gin.Context) {
	idKoor := c.GetString("id_koordinator")

	token, err := utils.GenerateQRToken(idKoor)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to generate QR token", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "QR Token generated", gin.H{"qr_token": token})
}

func DeactivateKiosk(c *gin.Context) {
	idKoor := c.GetString("id_koordinator")

	var koor models.Koordinator
	config.DB.Where("id = ?", idKoor).First(&koor)

	var user models.User
	config.DB.Where("id = ?", koor.UserID).First(&user)

	idAsisten := c.GetString("id_asisten")
	var asdosID *string
	if idAsisten != "" {
		asdosID = &idAsisten
	}

	// Re-issue token with is_kiosk_mode = false
	token, err := utils.GenerateToken(user.ID, user.Email, asdosID, &koor.ID, false)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to deactivate kiosk mode", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Kiosk Mode Deactivated", gin.H{"token": token})
}
