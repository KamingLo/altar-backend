package services

import (
	"altar/config"
	"altar/models"
	"errors"
	"fmt"
	"sort"
	"time"
)

// ─────────────────────────────────────────────
// DTO: SubstituteSessionInput (Create)
// ─────────────────────────────────────────────

type SubstituteSessionInput struct {
	IDSession      string  `json:"id_session"`
	IDRuangan      string  `json:"id_ruangan"`
	IDDosen        *string `json:"id_dosen"`
	IDAsdos1       *string `json:"id_asdos1"`
	IDAsdos2       *string `json:"id_asdos2"`
	SubstituteDate string  `json:"substitute_date"` // YYYY-MM-DD
	OriginalDate   string  `json:"original_date"`   // YYYY-MM-DD (cancelled regular date)
	SlotOption     int     `json:"slot_option"`     // 1–7
	Reason         string  `json:"reason"`
}

// ─────────────────────────────────────────────
// DTO: UpdateSubstituteStatusInput (Verify/Reject)
// ─────────────────────────────────────────────

type UpdateSubstituteStatusInput struct {
	Status            models.SubstituteSessionStatus `json:"status"`
	CoordinatorReason *string                        `json:"coordinator_reason"`
}

// ─────────────────────────────────────────────
// DTO: SubstituteSessionResponse
// ─────────────────────────────────────────────

type SubstituteSessionResponse struct {
	ID                string                         `json:"id"`
	Status            models.SubstituteSessionStatus `json:"status"`
	Reason            string                         `json:"reason"`
	CoordinatorReason *string                        `json:"coordinator_reason"`
	SubstituteDate    string                         `json:"substitute_date"`    // YYYY-MM-DD
	OriginalDate      string                         `json:"original_date"`      // YYYY-MM-DD
	TimeSlot          string                         `json:"time_slot"`          // "HH:mm – HH:mm"
	Room              string                         `json:"room"`               // "Name (Floor X)"
	IDDosen           *string                        `json:"id_dosen"`           // Added
	IDAsdos1          *string                        `json:"id_asdos1"`          // Added
	IDAsdos2          *string                        `json:"id_asdos2"`          // Added
	SubstituteTeacher string                         `json:"substitute_teacher"` // name of the replacement instructors (override)
	Session           *SessionResponse               `json:"session,omitempty"`
	CreatedAt         time.Time                      `json:"created_at"`
	UpdatedAt         time.Time                      `json:"updated_at"`
}

// ─────────────────────────────────────────────
// DTO: UnifiedJadwalResponse
// Used by GetScheduleByPeriod (Timeline Projection).
// Represents either a regular (REGULER) or substitute (PENGGANTI) session
// for a specific calendar date.
// ─────────────────────────────────────────────

type UnifiedJadwalResponse struct {
	IDSesi     string `json:"id_sesi"`
	Tipe       string `json:"tipe"`    // "REGULER" | "PENGGANTI"
	Tanggal    string `json:"tanggal"` // YYYY-MM-DD
	NamaKelas  string `json:"nama_kelas"`
	MataKuliah string `json:"mata_kuliah"`
	Ruangan    string `json:"ruangan"`  // "Nama Ruangan (Lantai X)"
	Pengajar   string `json:"pengajar"` // Nama dosen / "Asdos1 & Asdos2"
	Waktu      string `json:"waktu"`    // "HH:mm - HH:mm"
}

// ─────────────────────────────────────────────
// Internal: translateSubstituteSchedule
// Converts a concrete date string (YYYY-MM-DD) + slot option into
// two WIB time.Time values (KelasMulai, KelasBerakhir).
// ─────────────────────────────────────────────

func translateSubstituteSchedule(dateStr string, slotOption int) (time.Time, time.Time, error) {
	type timeSlot struct{ startH, startM, endH, endM int }
	slotMap := map[int]timeSlot{
		1: {7, 30, 9, 10},
		2: {9, 30, 11, 10},
		3: {11, 30, 13, 10},
		4: {13, 30, 15, 10},
		5: {15, 30, 17, 10},
		6: {17, 40, 19, 15},
		7: {19, 30, 21, 0},
	}
	slot, ok := slotMap[slotOption]
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid slot option: %d (valid: 1–7)", slotOption)
	}

	date, err := time.ParseInLocation("2006-01-02", dateStr, wib)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid substitute_date format, expected YYYY-MM-DD: %w", err)
	}

	startTime := time.Date(date.Year(), date.Month(), date.Day(), slot.startH, slot.startM, 0, 0, wib)
	endTime := time.Date(date.Year(), date.Month(), date.Day(), slot.endH, slot.endM, 0, 0, wib)
	return startTime, endTime, nil
}

// ─────────────────────────────────────────────
// Internal: checkSubstituteClash
// Detects room conflicts across both JadwalUtama and SubstituteSession.
// For JadwalUtama: checks weekly time-of-day overlap regardless of week.
// For SubstituteSession: checks only PENDING or VERIFIED entries (same room + exact date).
// ─────────────────────────────────────────────

func checkSubstituteClash(roomID string, startTime, endTime time.Time, excludeID string) error {
	db := config.DB
	var count int64

	// A. Conflict against existing main sessions (JadwalUtama).
	//    JadwalUtama stores anchor dates in January 2024 but the time-of-day is
	//    what matters for a weekly repeating schedule. We compare the week day
	//    and the time window by checking the stored kelas_mulai/kelas_berakhir
	//    times that fall on the same weekday AND overlap in clock time.
	weekday := int(startTime.Weekday()) // Go: 0=Sun, 1=Mon, ..., 6=Sat

	// Gunakan sintaks PostgreSQL: EXTRACT(DOW) dan casting ::time
	if err := db.Model(&models.JadwalUtama{}).
		Where("id_ruangan = ?", roomID).
		Where("EXTRACT(DOW FROM kelas_mulai) = ?", weekday).
		Where("kelas_mulai::time < ?::time AND kelas_berakhir::time > ?::time",
			endTime, startTime).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check main session room conflict: %w", err)
	}
	if count > 0 {
		return errors.New("room is already used by a regular session on this weekday and time slot")
	}

	// B. Conflict against other SubstituteSession entries (PENDING or VERIFIED only).
	query := db.Model(&models.SubstituteSession{}).
		Where("id_ruangan = ?", roomID).
		Where("status IN (?)", []models.SubstituteSessionStatus{models.StatusPending, models.StatusVerified}).
		Where("kelas_mulai < ? AND kelas_berakhir > ?", endTime, startTime)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	count = 0
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check substitute session room conflict: %w", err)
	}
	if count > 0 {
		return errors.New("room already has a pending or verified substitute session that overlaps with this time slot")
	}

	return nil
}

// ─────────────────────────────────────────────
// Internal: checkSubstituteTeacherClash
// Ensures a designated substitute teacher (asdos) has no conflicting
// schedule at the proposed substitute date and time slot.
//
// Two conflict sources are checked:
//   A. Regular sessions (JadwalUtama): same weekday + overlapping time.
//   B. Other SubstituteSessions (PENDING or VERIFIED): same exact date + overlapping time.
// ─────────────────────────────────────────────

func checkSubstituteTeacherClash(teacherIDs []string, startTime, endTime time.Time, excludeID string) error {
	if len(teacherIDs) == 0 {
		return nil
	}

	db := config.DB
	var count int64

	weekday := int(startTime.Weekday()) // Go: 0=Sun, 1=Mon, ..., 6=Sat

	// A. Conflict against JadwalUtama (regular sessions).
	//    Match on weekday (DOW) and overlapping clock time for ANY of the teachers.
	if err := db.Model(&models.JadwalUtama{}).
		Where("(id_dosen IN ? OR id_asdos1 IN ? OR id_asdos2 IN ?)", teacherIDs, teacherIDs, teacherIDs).
		Where("EXTRACT(DOW FROM kelas_mulai) = ?", weekday).
		Where("kelas_mulai::time < ?::time AND kelas_berakhir::time > ?::time", endTime, startTime).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check substitute teacher regular session conflict: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("one of the replacement teachers already has a regular class on this weekday and time slot")
	}

	// B. Conflict against other SubstituteSessions (PENDING or VERIFIED).
	//    The teachers may appear in another substitute request.
	query := db.Model(&models.SubstituteSession{}).
		Where("(id_dosen IN ? OR id_asdos1 IN ? OR id_asdos2 IN ?)", teacherIDs, teacherIDs, teacherIDs).
		Where("status IN (?)", []models.SubstituteSessionStatus{models.StatusPending, models.StatusVerified}).
		Where("kelas_mulai < ? AND kelas_berakhir > ?", endTime, startTime)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	count = 0
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check substitute teacher session conflict: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("one of the replacement teachers already has a teaching schedule at that time")
	}

	return nil
}

// ─────────────────────────────────────────────
// Internal: buildSubstituteResponse
// Builds the response DTO from a fully preloaded SubstituteSession.
// ─────────────────────────────────────────────

func buildSubstituteResponse(sub *models.SubstituteSession) SubstituteSessionResponse {
	roomStr := ""
	if sub.Ruangan != nil {
		roomStr = fmt.Sprintf("%s (Lantai %d)", sub.Ruangan.NamaRuangan, sub.Ruangan.Lantai)
	}

	timeSlot := fmt.Sprintf("%s – %s",
		sub.KelasMulai.In(wib).Format("15:04"),
		sub.KelasBerakhir.In(wib).Format("15:04"),
	)

	dateStr := sub.SubstituteDate.In(wib).Format("2006-01-02")
	originalDateStr := sub.OriginalDate.In(wib).Format("2006-01-02")

	// Resolve substitute teacher name (empty string when none is assigned)
	substituteTeacherName := ""
	if sub.Session != nil {
		substituteTeacherName = FormatInstructor(sub.Session, sub)
	}

	var sessionResp *SessionResponse
	if sub.Session != nil {
		r, err := loadSessionRelations(sub.Session)
		if err == nil {
			sessionResp = &r
		}
	}

	return SubstituteSessionResponse{
		ID:                sub.ID,
		Status:            sub.Status,
		Reason:            sub.Reason,
		CoordinatorReason: sub.CoordinatorReason,
		SubstituteDate:    dateStr,
		OriginalDate:      originalDateStr,
		TimeSlot:          timeSlot,
		Room:              roomStr,
		IDDosen:           sub.IDDosen,
		IDAsdos1:          sub.IDAsdos1,
		IDAsdos2:          sub.IDAsdos2,
		SubstituteTeacher: substituteTeacherName,
		Session:           sessionResp,
		CreatedAt:         sub.CreatedAt,
		UpdatedAt:         sub.UpdatedAt,
	}
}

// ─────────────────────────────────────────────
// Internal: FormatInstructor
// Returns the instructor display string.
// If sub is provided and has any override instructors (Dosen/Asdos1/Asdos2),
// their names are returned with "(Pengganti)" appended.
// Otherwise, it falls back to the original session instructors.
// ─────────────────────────────────────────────

func FormatInstructor(session *models.JadwalUtama, sub *models.SubstituteSession) string {
	// If a substitute session is provided and has at least one override instructor:
	if sub != nil && (sub.IDDosen != nil || sub.IDAsdos1 != nil || sub.IDAsdos2 != nil) {
		var names []string
		if sub.Dosen != nil {
			names = append(names, sub.Dosen.Nama)
		}
		if sub.Asdos1 != nil {
			names = append(names, sub.Asdos1.User.Username)
		}
		if sub.Asdos2 != nil {
			names = append(names, sub.Asdos2.User.Username)
		}

		if len(names) == 0 {
			return "-"
		}

		instructorStr := ""
		if len(names) == 1 {
			instructorStr = names[0]
		} else {
			// Join all but the last with ", ", then join the last with " & "
			instructorStr = fmt.Sprintf("%s & %s", names[0], names[1])
			if len(names) == 3 {
				instructorStr = fmt.Sprintf("%s, %s & %s", names[0], names[1], names[2])
			}
		}
		return fmt.Sprintf("%s (Pengganti)", instructorStr)
	}

	// Fall back to the original session instructor.
	var names []string
	if session.Dosen != nil {
		names = append(names, session.Dosen.Nama)
	}
	if session.Asdos1 != nil {
		names = append(names, session.Asdos1.User.Username)
	}
	if session.Asdos2 != nil {
		names = append(names, session.Asdos2.User.Username)
	}

	if len(names) == 0 {
		return "-"
	}

	if len(names) == 1 {
		return names[0]
	}

	if len(names) == 2 {
		return fmt.Sprintf("%s & %s", names[0], names[1])
	}

	return fmt.Sprintf("%s, %s & %s", names[0], names[1], names[2])
}

// ─────────────────────────────────────────────
// Internal: buildUnifiedFromRegular
// Converts a JadwalUtama into a UnifiedJadwalResponse for a specific calendar date.
// ─────────────────────────────────────────────

func buildUnifiedFromRegular(session *models.JadwalUtama, date time.Time) UnifiedJadwalResponse {
	ruanganStr := fmt.Sprintf("%s (Lantai %d)", session.Ruangan.NamaRuangan, session.Ruangan.Lantai)

	waktu := fmt.Sprintf("%s - %s",
		session.KelasMulai.In(wib).Format("15:04"),
		session.KelasBerakhir.In(wib).Format("15:04"),
	)

	return UnifiedJadwalResponse{
		IDSesi:     session.ID,
		Tipe:       "REGULER",
		Tanggal:    date.Format("2006-01-02"),
		NamaKelas:  session.Kelas.NamaKelas,
		MataKuliah: session.MataKuliah.NamaMK,
		Ruangan:    ruanganStr,
		Pengajar:   FormatInstructor(session, nil), // regular session: no substitute teacher
		Waktu:      waktu,
	}
}

// ─────────────────────────────────────────────
// Internal: buildUnifiedFromSubstitute
// Converts a VERIFIED SubstituteSession into a UnifiedJadwalResponse.
// Uses formatInstructor: if a SubstituteTeacher is set, their name is shown
// with "(Substitute Teacher)"; otherwise the original session instructor is used.
// ─────────────────────────────────────────────

func buildUnifiedFromSubstitute(sub *models.SubstituteSession) UnifiedJadwalResponse {
	ruanganStr := ""
	if sub.Ruangan != nil {
		ruanganStr = fmt.Sprintf("%s (Lantai %d)", sub.Ruangan.NamaRuangan, sub.Ruangan.Lantai)
	}

	// Resolve instructor: substitute formation takes priority over the original one.
	instructorStr := "-"
	if sub.Session != nil {
		instructorStr = FormatInstructor(sub.Session, sub)
	}

	namaKelas := ""
	mataKuliah := ""
	if sub.Session != nil {
		namaKelas = sub.Session.Kelas.NamaKelas
		mataKuliah = sub.Session.MataKuliah.NamaMK
	}

	waktu := fmt.Sprintf("%s - %s",
		sub.KelasMulai.In(wib).Format("15:04"),
		sub.KelasBerakhir.In(wib).Format("15:04"),
	)

	return UnifiedJadwalResponse{
		IDSesi:     sub.ID,
		Tipe:       "PENGGANTI",
		Tanggal:    sub.SubstituteDate.In(wib).Format("2006-01-02"),
		NamaKelas:  namaKelas,
		MataKuliah: mataKuliah,
		Ruangan:    ruanganStr,
		Pengajar:   instructorStr,
		Waktu:      waktu,
	}
}

// ─────────────────────────────────────────────
// Service: CreateSubstituteSession
// Validates weekday alignment, room availability, then saves a new PENDING substitute request.
// ─────────────────────────────────────────────

func CreateSubstituteSession(input *SubstituteSessionInput) (SubstituteSessionResponse, error) {
	db := config.DB

	// 1. Ensure the referenced session exists
	var session models.JadwalUtama
	if err := db.First(&session, "id = ?", input.IDSession).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("session with ID %s not found", input.IDSession)
	}

	// 2. Ensure the room exists
	var ruangan models.Ruangan
	if err := db.First(&ruangan, "id = ?", input.IDRuangan).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("room with ID %s not found", input.IDRuangan)
	}

	// 3. Validate substitute instructors exist (when provided)
	var teacherIDs []string

	if input.IDDosen != nil && *input.IDDosen != "" {
		if err := db.First(&models.Dosen{}, "id = ?", *input.IDDosen).Error; err != nil {
			return SubstituteSessionResponse{}, fmt.Errorf("dosen pengganti with ID %s not found", *input.IDDosen)
		}
		teacherIDs = append(teacherIDs, *input.IDDosen)
	} else {
		input.IDDosen = nil
	}

	if input.IDAsdos1 != nil && *input.IDAsdos1 != "" {
		if err := db.First(&models.AsistenDosen{}, "id = ?", *input.IDAsdos1).Error; err != nil {
			return SubstituteSessionResponse{}, fmt.Errorf("asdos pengganti 1 with ID %s not found", *input.IDAsdos1)
		}
		teacherIDs = append(teacherIDs, *input.IDAsdos1)
	} else {
		input.IDAsdos1 = nil
	}

	if input.IDAsdos2 != nil && *input.IDAsdos2 != "" {
		if err := db.First(&models.AsistenDosen{}, "id = ?", *input.IDAsdos2).Error; err != nil {
			return SubstituteSessionResponse{}, fmt.Errorf("asdos pengganti 2 with ID %s not found", *input.IDAsdos2)
		}
		teacherIDs = append(teacherIDs, *input.IDAsdos2)
	} else {
		input.IDAsdos2 = nil
	}

	// 4. Parse original_date and validate weekday matches the regular session's weekday
	originalDate, err := time.Parse("2006-01-02", input.OriginalDate)
	if err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("invalid original_date format, expected YYYY-MM-DD: %w", err)
	}

	regularWeekday := session.KelasMulai.In(wib).Weekday()
	originalWeekday := originalDate.Weekday()
	if originalWeekday != regularWeekday {
		return SubstituteSessionResponse{}, fmt.Errorf(
			"original_date weekday (%s) does not match the regular session weekday (%s)",
			originalWeekday.String(), regularWeekday.String(),
		)
	}

	// 5. Translate date + slot into concrete time.Time values
	startTime, endTime, err := translateSubstituteSchedule(input.SubstituteDate, input.SlotOption)
	if err != nil {
		return SubstituteSessionResponse{}, err
	}

	// 6. Room clash detection against JadwalUtama and other SubstituteSessions
	if err := checkSubstituteClash(input.IDRuangan, startTime, endTime, ""); err != nil {
		return SubstituteSessionResponse{}, err
	}

	// 7. Teacher clash detection (for all override instructors)
	if err := checkSubstituteTeacherClash(teacherIDs, startTime, endTime, ""); err != nil {
		return SubstituteSessionResponse{}, err
	}

	// 8. Parse substitute date (date-only, midnight WIB)
	substituteDate, _ := time.Parse("2006-01-02", input.SubstituteDate)

	// 9. Persist
	sub := models.SubstituteSession{
		IDSession:      input.IDSession,
		IDRuangan:      input.IDRuangan,
		IDDosen:        input.IDDosen,
		IDAsdos1:       input.IDAsdos1,
		IDAsdos2:       input.IDAsdos2,
		Reason:         input.Reason,
		Status:         models.StatusPending,
		SubstituteDate: substituteDate,
		OriginalDate:   originalDate,
		KelasMulai:     startTime,
		KelasBerakhir:  endTime,
	}
	if err := db.Create(&sub).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to create substitute session: %w", err)
	}

	// 10. Reload with full preloads for the response
	if err := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").
		Preload("Ruangan").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User").
		First(&sub, "id = ?", sub.ID).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to reload substitute session: %w", err)
	}

	return buildSubstituteResponse(&sub), nil
}

// ─────────────────────────────────────────────
// Service: GetAllSubstituteSessions
// Returns all substitute sessions, optionally filtered by status.
// ─────────────────────────────────────────────

func GetAllSubstituteSessions(statusFilter string) ([]SubstituteSessionResponse, error) {
	db := config.DB
	var subs []models.SubstituteSession

	query := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").
		Preload("Ruangan").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User")

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Order("created_at DESC").Find(&subs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch substitute sessions: %w", err)
	}

	responses := make([]SubstituteSessionResponse, 0, len(subs))
	for i := range subs {
		responses = append(responses, buildSubstituteResponse(&subs[i]))
	}
	return responses, nil
}

// ─────────────────────────────────────────────
// Service: GetSubstituteSessionByID
// ─────────────────────────────────────────────

func GetSubstituteSessionByID(id string) (SubstituteSessionResponse, error) {
	db := config.DB
	var sub models.SubstituteSession

	if err := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").
		Preload("Ruangan").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User").
		First(&sub, "id = ?", id).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("substitute session not found: %w", err)
	}

	return buildSubstituteResponse(&sub), nil
}

// ─────────────────────────────────────────────
// Service: DeleteSubstituteSession
// Soft deletes a substitute session (only if status is PENDING).
// ─────────────────────────────────────────────

func DeleteSubstituteSession(id string) error {
	db := config.DB
	var sub models.SubstituteSession

	if err := db.First(&sub, "id = ?", id).Error; err != nil {
		return fmt.Errorf("substitute session not found: %w", err)
	}

	if sub.Status != models.StatusPending {
		return fmt.Errorf("cannot delete substitute session with status %s, only PENDING is allowed", sub.Status)
	}

	if err := db.Delete(&sub).Error; err != nil {
		return fmt.Errorf("failed to delete substitute session: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────
// Service: UpdateSubstituteStatus
// Coordinator approves (VERIFIED) or rejects (REJECTED) a pending request.
// ─────────────────────────────────────────────

func UpdateSubstituteStatus(id string, input *UpdateSubstituteStatusInput) (SubstituteSessionResponse, error) {
	db := config.DB

	// 1. Validate allowed transitions
	if input.Status != models.StatusVerified && input.Status != models.StatusRejected {
		return SubstituteSessionResponse{}, errors.New("status must be either 'VERIFIED' or 'REJECTED'")
	}

	// 2. Cek eksistensi dan status saat ini
	var sub models.SubstituteSession
	if err := db.First(&sub, "id = ?", id).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("substitute session with ID %s not found", id)
	}

	if sub.Status == models.StatusVerified || sub.Status == models.StatusRejected {
		return SubstituteSessionResponse{}, fmt.Errorf(
			"substitute session is already finalised with status '%s' and cannot be changed", sub.Status,
		)
	}

	// 3. THE FIX: Update status dan optional coordinator_reason menggunakan map (mencegah GORM update relasi)
	updates := map[string]interface{}{
		"status": input.Status,
	}
	if input.CoordinatorReason != nil {
		updates["coordinator_reason"] = *input.CoordinatorReason
	}

	if err := db.Model(&models.SubstituteSession{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to update substitute session status: %w", err)
	}

	// 4. Reload data secara utuh beserta relasinya untuk response JSON
	if err := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").
		Preload("Ruangan").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User").
		First(&sub, "id = ?", id).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to reload substitute session after update: %w", err)
	}

	return buildSubstituteResponse(&sub), nil
}

// ─────────────────────────────────────────────
// Service: GetTimelineJadwal (Proyeksi Kalender Pintar)
//
// Mengelompokkan hasil akhir berdasarkan list harian yang diurutkan
// berdasarkan tanggal dan jam mulai kelas.
// ─────────────────────────────────────────────

func GetTimelineJadwal(startDateStr, endDateStr, idSemester, asdosID string) ([]UnifiedJadwalResponse, error) {
	db := config.DB

	// 1. Parse Range: Generate list semua tanggal (tipe time.Time) dari start_date sampai end_date
	startDate, err := time.ParseInLocation("2006-01-02", startDateStr, wib)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}
	endDate, err := time.ParseInLocation("2006-01-02", endDateStr, wib)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}
	if endDate.Before(startDate) {
		return nil, errors.New("end_date must be greater than or equal to start_date")
	}
	if idSemester == "" {
		return nil, errors.New("id_semester is required")
	}

	// 2. Query Data Reguler: Tarik semua JadwalUtama berdasarkan id_semester
	//    Jika asdosID tidak kosong, filter id_asdos1 = asdosID OR id_asdos2 = asdosID.
	var regularSessions []models.JadwalUtama
	regularQuery := db.
		Preload("Kelas").
		Preload("MataKuliah").
		Preload("Ruangan").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User").
		Where("id_semester = ?", idSemester)
	if asdosID != "" {
		regularQuery = regularQuery.Where("id_asdos1 = ? OR id_asdos2 = ?", asdosID, asdosID)
	}
	if err := regularQuery.Find(&regularSessions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch regular sessions: %w", err)
	}

	// 3. Query Data Pengganti: Tarik semua SubstituteSession yang statusnya VERIFIED
	//    dan bersinggungan dengan range tanggal tersebut.
	var substituteSessions []models.SubstituteSession
	substituteQuery := db.
		Preload("Session").
		Preload("Session.Kelas").
		Preload("Session.MataKuliah").
		Preload("Session.Dosen").
		Preload("Session.Asdos1").Preload("Session.Asdos1.User").
		Preload("Session.Asdos2").Preload("Session.Asdos2.User").
		Preload("Ruangan").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User").
		Where("status = ?", models.StatusVerified).
		Where("substitute_date::date >= ? AND substitute_date::date <= ?",
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if asdosID != "" {
		// Logika Asdos: asdos_pengganti (override) = asdosID OR (asdos_pengganti_all_nil AND (session.id_asdos1 = asdosID OR session.id_asdos2 = asdosID))
		substituteQuery = substituteQuery.Joins("JOIN jadwal_utamas ju ON ju.id = substitute_sessions.id_session").
			Where(
				"(substitute_sessions.id_asdos1 = ? OR substitute_sessions.id_asdos2 = ?) OR "+
					"((substitute_sessions.id_dosen IS NULL AND substitute_sessions.id_asdos1 IS NULL AND substitute_sessions.id_asdos2 IS NULL) AND (ju.id_asdos1 = ? OR ju.id_asdos2 = ?))",
				asdosID, asdosID, asdosID, asdosID,
			)
	}
	if err := substituteQuery.Find(&substituteSessions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch substitute sessions: %w", err)
	}

	// 4. Logika Blacklist (Tanggal Batal): Buat Hash Map map[string]bool
	//    Format key: YYYY-MM-DD_IDSession.
	blacklist := make(map[string]bool, len(substituteSessions))
	for i := range substituteSessions {
		sub := &substituteSessions[i]
		key := sub.OriginalDate.In(wib).Format("2006-01-02") + "_" + sub.IDSession
		blacklist[key] = true
	}

	results := make([]UnifiedJadwalResponse, 0)

	// 5. Logika Projection (Looping Harian):
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		dayWeekday := day.Weekday()
		dateStr := day.Format("2006-01-02")

		// A. Render Reguler: Cek semua JadwalUtama
		for i := range regularSessions {
			sess := &regularSessions[i]
			// Jika weekday jadwal reguler SAMA dengan weekday tanggal iterasi saat ini:
			if sess.KelasMulai.In(wib).Weekday() != dayWeekday {
				continue
			}

			// Cek Blacklist: Jika kombinasi Tanggal Iterasi + ID JadwalUtama ADA di Hash Map
			blacklistKey := dateStr + "_" + sess.ID
			if blacklist[blacklistKey] {
				// SKIP jadwal ini (karena hari itu libur/dipindah)
				continue
			}

			// Jika TIDAK ADA di Blacklist, masukkan ke response dengan tipe "REGULAR"
			unified := buildUnifiedFromRegular(sess, day)
			unified.Tipe = "REGULAR" // memastikan format tepat sesuai req
			results = append(results, unified)
		}

		// B. Render Pengganti: Cek semua SubstituteSession
		for i := range substituteSessions {
			sub := &substituteSessions[i]
			// Jika SubstituteDate SAMA dengan tanggal iterasi saat ini
			if sub.SubstituteDate.In(wib).Format("2006-01-02") == dateStr {
				// Masukkan ke response dengan tipe "SUBSTITUTE".
				// (Fungsi buildUnifiedFromSubstitute sudah menangani logic AsdosPengganti)
				unified := buildUnifiedFromSubstitute(sub)
				unified.Tipe = "SUBSTITUTE"
				results = append(results, unified)
			}
		}
	}

	// 6. Kelompokkan hasil akhir (List harian diurutkan berdasarkan jam mulai kelas)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Tanggal != results[j].Tanggal {
			return results[i].Tanggal < results[j].Tanggal
		}
		return results[i].Waktu < results[j].Waktu
	})

	return results, nil
}

// ─────────────────────────────────────────────
// Service: GetDailyAsdosSessions
// Daily view for an Asdos for a specific date (combining regular & substitute).
// ─────────────────────────────────────────────
