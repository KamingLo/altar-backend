package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// --- User Handlers ---

func CreateUser(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := services.CreateUser(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "User created successfully", nil)
}

func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	search := c.Query("search")

	users, err := services.GetAllUsers(page, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", users)
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := services.GetUserByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "User not found", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := services.UpdateUser(id, input.Username, input.Email); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "User updated successfully", nil)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteUser(id); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "User deleted successfully", nil)
}

// --- Asisten Dosen Handlers ---

func CreateAsdos(c *gin.Context) {
	var input struct {
		UserID      string `json:"user_id" binding:"required"`
		NIM         string `json:"nim" binding:"required"`
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Data is not complete", err)
		return
	}

	if err := services.CreateAsdos(input.UserID, input.NIM, input.PhoneNumber); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Asisten Dosen created successfully", nil)
}

func GetAllAsdos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	search := c.Query("search")

	asdos, err := services.GetAllAsdos(page, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch data", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", asdos)
}

func GetAsdosByID(c *gin.Context) {
	id := c.Param("id")
	asdos, err := services.GetAsdosByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Asisten Dosen not found", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", asdos)
}

func UpdateAsdos(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NIM         string `json:"nim" binding:"required"`
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := services.UpdateAsdos(id, input.NIM, input.PhoneNumber); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Asisten Dosen updated successfully", nil)
}

func DeleteAsdos(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteAsdos(id); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Asisten Dosen deleted successfully", nil)
}

// --- Koordinator Handlers ---

func CreateKoordinator(c *gin.Context) {
	var input struct {
		UserID string `json:"user_id" binding:"required"`
		NIP    string `json:"nip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Data is not complete", err)
		return
	}

	if err := services.CreateKoordinator(input.UserID, input.NIP); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Koordinator created successfully", nil)
}

func GetAllKoordinator(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	search := c.Query("search")

	koor, err := services.GetAllKoordinator(page, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch data", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", koor)
}

func GetKoordinatorByID(c *gin.Context) {
	id := c.Param("id")
	koor, err := services.GetKoordinatorByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Koordinator not found", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", koor)
}

func UpdateKoordinator(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NIP string `json:"nip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := services.UpdateKoordinator(id, input.NIP); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Koordinator updated successfully", nil)
}

func DeleteKoordinator(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteKoordinator(id); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Koordinator deleted successfully", nil)
}
