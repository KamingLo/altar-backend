package controllers

import (
	"altar/services"
	"altar/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// getCutoffDates menghitung startDate dan endDate secara dinamis
// menggunakan siklus cut-off tanggal 25.
func getCutoffDates() (time.Time, time.Time) {
	now := time.Now()
	year, month, day := now.Date()

	var startDate, endDate time.Time
	if day <= 25 {
		// Siklus bulan ini: 26 bulan lalu hingga 25 bulan ini
		startDate = time.Date(year, month-1, 26, 0, 0, 0, 0, now.Location())
		endDate = time.Date(year, month, 25, 0, 0, 0, 0, now.Location())
	} else {
		// Siklus bulan depan: 26 bulan ini hingga 25 bulan depan
		startDate = time.Date(year, month, 26, 0, 0, 0, 0, now.Location())
		endDate = time.Date(year, month+1, 25, 0, 0, 0, 0, now.Location())
	}
	return startDate, endDate
}

func GetRekapPresensiMe(c *gin.Context) {
	asdosID := c.GetString("id_asisten")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	// Jika parameter tanggal kosong, gunakan logika cut-off dinamis (tanggal 25)
	if startDateStr == "" || endDateStr == "" {
		startDate, endDate = getCutoffDates()
	} else {
		// Jika terisi, lakukan parsing tanggal seperti biasa
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD", nil)
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD", nil)
			return
		}
	}

	res, err := services.GetRekapPresensi(asdosID, startDate, endDate)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to calculate attendance recap", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Attendance recap fetched successfully", res)
}

func GetRekapPresensi(c *gin.Context) {
	asdosID := c.Query("asdos_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	// Jika parameter tanggal kosong, gunakan logika cut-off dinamis (tanggal 25)
	if startDateStr == "" || endDateStr == "" {
		startDate, endDate = getCutoffDates()
	} else {
		// Jika terisi, lakukan parsing tanggal seperti biasa
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD", nil)
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD", nil)
			return
		}
	}

	// Panggil logika service yang sesuai berdasarkan keberadaan parameter asdosID
	if asdosID == "" {
		res, err := services.GetAllRekapPresensi(startDate, endDate)
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Failed to calculate all attendance recaps", err)
			return
		}
		utils.SendSuccess(c, http.StatusOK, "All attendance recaps fetched successfully", res)
	} else {
		res, err := services.GetRekapPresensi(asdosID, startDate, endDate)
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Failed to calculate attendance recap", err)
			return
		}
		utils.SendSuccess(c, http.StatusOK, "Attendance recap fetched successfully", res)
	}
}
