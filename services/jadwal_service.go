package services

import (
	"altar/config"
	"altar/models"
	"errors"
	"fmt"
	"log"
	"time"
)

// ─────────────────────────────────────────────
// Timezone
// ─────────────────────────────────────────────

var wib = time.FixedZone("WIB", 7*60*60)

// ─────────────────────────────────────────────
// Helper: TranslateSchedule
// Converts opsi_hari + opsi_jam integers into WIB time.Time values
// using the reference week of the first week of January 2024.
//
// Day mapping  (opsi_hari): 1=Mon(Jan 1), 2=Tue(Jan 2), …, 6=Sat(Jan 6)
// Slot mapping (opsi_jam) : 1=07:30-09:10, 2=09:30-11:10, 3=11:30-13:10,
//                           4=13:30-15:10, 5=15:30-17:10, 6=17:40-19:15,
//                           7=19:30-21:00
// ─────────────────────────────────────────────

func TranslateSchedule(dayOption int, slotOption int) (startTime time.Time, endTime time.Time, err error) {
	dayToDate := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6}
	date, ok := dayToDate[dayOption]
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid day option: %d (valid: 1–6)", dayOption)
	}

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

	startTime = time.Date(2024, time.January, date, slot.startH, slot.startM, 0, 0, wib)
	endTime = time.Date(2024, time.January, date, slot.endH, slot.endM, 0, 0, wib)
	return startTime, endTime, nil
}

// ─────────────────────────────────────────────
// DTO: Input for Create & Update
// ─────────────────────────────────────────────

type SessionInput struct {
	IDKelas    string  `json:"id_kelas"`
	IDMk       string  `json:"id_mk"`
	IDSemester string  `json:"id_semester"`
	IDRuangan  string  `json:"id_ruangan"`
	IDAsdos1   *string `json:"id_asdos1"`
	IDAsdos2   *string `json:"id_asdos2"`
	IDDosen    *string `json:"id_dosen"`
	DayOption  int     `json:"opsi_hari"`
	SlotOption int     `json:"opsi_jam"`
}

// ─────────────────────────────────────────────
// DTO: Custom Response
// ─────────────────────────────────────────────

type SessionResponse struct {
	ID         string `json:"id_sesi"`
	NamaKelas  string `json:"nama_kelas"`
	MataKuliah string `json:"mata_kuliah"`
	Ruangan    string `json:"ruangan"`  // Format: "Room Name (Floor X)"
	Pengajar   string `json:"pengajar"` // Format: Lecturer name OR "Asdos1 & Asdos2" OR "Asdos1"
	Waktu      string `json:"waktu"`    // Format: "Senin, 07:30 - 09:10"
}

// DailySessionResponse adalah gabungan antara response standar
// dengan label tipe untuk membedakan jadwal reguler dan pengganti.
type DailySessionResponse struct {
	SessionResponse
	TipeJadwal string `json:"tipe_jadwal"` // "REGULAR" atau "PENGGANTI"
}

// ─────────────────────────────────────────────
// Internal: buildSessionResponse
// Maps raw model + preloaded relations into a SessionResponse.
// ─────────────────────────────────────────────

func buildSessionResponse(
	session *models.JadwalUtama,
	kelas *models.Kelas,
	mk *models.MataKuliah,
	ruangan *models.Ruangan,
	dosen *models.Dosen,
	asdos1 *models.AsistenDosen,
	asdos2 *models.AsistenDosen,
) SessionResponse {
	// "Room Name (Floor X)"
	roomStr := fmt.Sprintf("%s (Lantai %d)", ruangan.NamaRuangan, ruangan.Lantai)

	// Instructor string: Lecturer > Asdos1 & Asdos2 > Asdos1 only
	var instructorStr string
	switch {
	case dosen != nil:
		instructorStr = dosen.Nama
	case asdos1 != nil && asdos2 != nil:
		instructorStr = fmt.Sprintf("%s & %s", asdos1.User.Username, asdos2.User.Username)
	case asdos1 != nil:
		instructorStr = asdos1.User.Username
	default:
		instructorStr = "-"
	}

	// "Senin, 07:30 - 09:10"
	dayNames := map[int]string{1: "Senin", 2: "Selasa", 3: "Rabu", 4: "Kamis", 5: "Jumat", 6: "Sabtu"}
	dayName := dayNames[session.KelasMulai.In(wib).Day()]
	scheduleStr := fmt.Sprintf("%s, %s - %s",
		dayName,
		session.KelasMulai.In(wib).Format("15:04"),
		session.KelasBerakhir.In(wib).Format("15:04"),
	)

	return SessionResponse{
		ID:         session.ID,
		NamaKelas:  kelas.NamaKelas,
		MataKuliah: mk.NamaMK,
		Ruangan:    roomStr,
		Pengajar:   instructorStr,
		Waktu:      scheduleStr,
	}
}

// ─────────────────────────────────────────────
// Internal: loadSessionRelations
// Fetches all related records from DB and builds a SessionResponse.
// ─────────────────────────────────────────────

func loadSessionRelations(session *models.JadwalUtama) (SessionResponse, error) {
	db := config.DB

	var kelas models.Kelas
	if err := db.First(&kelas, "id = ?", session.IDKelas).Error; err != nil {
		return SessionResponse{}, fmt.Errorf("class not found: %w", err)
	}

	var mk models.MataKuliah
	if err := db.First(&mk, "id = ?", session.IDMk).Error; err != nil {
		return SessionResponse{}, fmt.Errorf("course not found: %w", err)
	}

	var ruangan models.Ruangan
	if err := db.First(&ruangan, "id = ?", session.IDRuangan).Error; err != nil {
		return SessionResponse{}, fmt.Errorf("room not found: %w", err)
	}

	var dosen *models.Dosen
	if session.IDDosen != nil {
		d := &models.Dosen{}
		if err := db.First(d, "id = ?", *session.IDDosen).Error; err != nil {
			return SessionResponse{}, fmt.Errorf("lecturer not found: %w", err)
		}
		dosen = d
	}

	var asdos1 *models.AsistenDosen
	if session.IDAsdos1 != nil {
		a := &models.AsistenDosen{}
		if err := db.Preload("User").First(a, "id = ?", *session.IDAsdos1).Error; err != nil {
			return SessionResponse{}, fmt.Errorf("assistant lecturer 1 not found: %w", err)
		}
		asdos1 = a
	}

	var asdos2 *models.AsistenDosen
	if session.IDAsdos2 != nil {
		a := &models.AsistenDosen{}
		if err := db.Preload("User").First(a, "id = ?", *session.IDAsdos2).Error; err != nil {
			return SessionResponse{}, fmt.Errorf("assistant lecturer 2 not found: %w", err)
		}
		asdos2 = a
	}

	return buildSessionResponse(session, &kelas, &mk, &ruangan, dosen, asdos1, asdos2), nil
}

// ─────────────────────────────────────────────
// Internal: validateInstructorXOR
// Enforces mutual exclusivity between Dosen and Asdos (Rule 2).
// ─────────────────────────────────────────────

func validateInstructorXOR(input *SessionInput) error {
	hasDosen := input.IDDosen != nil && *input.IDDosen != ""
	hasAsdos1 := input.IDAsdos1 != nil && *input.IDAsdos1 != ""
	hasAsdos2 := input.IDAsdos2 != nil && *input.IDAsdos2 != ""

	if !hasDosen && !hasAsdos1 && !hasAsdos2 {
		return errors.New("instructor is required: provide id_dosen OR id_asdos1 (and optionally id_asdos2)")
	}
	if hasDosen && (hasAsdos1 || hasAsdos2) {
		return errors.New("instructor conflict: id_dosen cannot be set together with id_asdos1 or id_asdos2")
	}
	if hasAsdos2 && !hasAsdos1 {
		return errors.New("id_asdos2 cannot be set without id_asdos1")
	}
	return nil
}

// ─────────────────────────────────────────────
// Internal: checkScheduleClash
// Detects time conflicts for room, lecturer, and assistants,
// scoped strictly to the given semester (Rule 3).
// excludeID is the session ID to exclude during Update checks.
// ─────────────────────────────────────────────

func checkScheduleClash(input *SessionInput, startTime, endTime time.Time, excludeID string) error {
	db := config.DB
	var count int64

	base := db.Model(&models.JadwalUtama{}).Where("id_semester = ?", input.IDSemester)
	if excludeID != "" {
		base = base.Where("id != ?", excludeID)
	}

	// A. Room conflict
	if err := base.Where(
		"id_ruangan = ? AND (kelas_mulai < ? AND kelas_berakhir > ?)",
		input.IDRuangan, endTime, startTime,
	).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check room conflict: %w", err)
	}
	if count > 0 {
		return errors.New("room is already occupied at this time slot in the current semester")
	}

	// B. Lecturer conflict
	if input.IDDosen != nil && *input.IDDosen != "" {
		count = 0
		if err := base.Where(
			"id_dosen = ? AND (kelas_mulai < ? AND kelas_berakhir > ?)",
			*input.IDDosen, endTime, startTime,
		).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check lecturer conflict: %w", err)
		}
		if count > 0 {
			return errors.New("lecturer already has another session at this time in the current semester")
		}
	}

	// C. Assistant lecturer 1 conflict
	if input.IDAsdos1 != nil && *input.IDAsdos1 != "" {
		count = 0
		if err := base.Where(
			"(id_asdos1 = ? OR id_asdos2 = ?) AND (kelas_mulai < ? AND kelas_berakhir > ?)",
			*input.IDAsdos1, *input.IDAsdos1, endTime, startTime,
		).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check assistant lecturer 1 conflict: %w", err)
		}
		if count > 0 {
			return errors.New("assistant lecturer 1 already has another session at this time in the current semester")
		}
	}

	// D. Assistant lecturer 2 conflict
	if input.IDAsdos2 != nil && *input.IDAsdos2 != "" {
		count = 0
		if err := base.Where(
			"(id_asdos1 = ? OR id_asdos2 = ?) AND (kelas_mulai < ? AND kelas_berakhir > ?)",
			*input.IDAsdos2, *input.IDAsdos2, endTime, startTime,
		).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check assistant lecturer 2 conflict: %w", err)
		}
		if count > 0 {
			return errors.New("assistant lecturer 2 already has another session at this time in the current semester")
		}
	}

	return nil
}

// ─────────────────────────────────────────────
// Internal: normalizeSessionInput
// Converts empty string pointers to nil.
// ─────────────────────────────────────────────

func normalizeSessionInput(input *SessionInput) {
	if input.IDDosen != nil && *input.IDDosen == "" {
		input.IDDosen = nil
	}
	if input.IDAsdos1 != nil && *input.IDAsdos1 == "" {
		input.IDAsdos1 = nil
	}
	if input.IDAsdos2 != nil && *input.IDAsdos2 == "" {
		input.IDAsdos2 = nil
	}
}

// ─────────────────────────────────────────────
// Service: CreateSession
// ─────────────────────────────────────────────

func CreateSession(input *SessionInput) (SessionResponse, error) {
	normalizeSessionInput(input)

	if err := validateInstructorXOR(input); err != nil {
		return SessionResponse{}, err
	}

	startTime, endTime, err := TranslateSchedule(input.DayOption, input.SlotOption)
	if err != nil {
		return SessionResponse{}, err
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate mandatory entity existence
	if err := tx.First(&models.Kelas{}, "id = ?", input.IDKelas).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("class with ID %s not found", input.IDKelas)
	}
	if err := tx.First(&models.MataKuliah{}, "id = ?", input.IDMk).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("course with ID %s not found", input.IDMk)
	}
	if err := tx.First(&models.Semester{}, "id = ?", input.IDSemester).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("semester with ID %s not found", input.IDSemester)
	}
	if err := tx.First(&models.Ruangan{}, "id = ?", input.IDRuangan).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("room with ID %s not found", input.IDRuangan)
	}

	// Validate optional entity existence
	if input.IDDosen != nil {
		if err := tx.First(&models.Dosen{}, "id = ?", *input.IDDosen).Error; err != nil {
			tx.Rollback()
			return SessionResponse{}, fmt.Errorf("lecturer with ID %s not found", *input.IDDosen)
		}
	}
	if input.IDAsdos1 != nil {
		if err := tx.First(&models.AsistenDosen{}, "id = ?", *input.IDAsdos1).Error; err != nil {
			tx.Rollback()
			return SessionResponse{}, fmt.Errorf("assistant lecturer 1 with ID %s not found", *input.IDAsdos1)
		}
	}
	if input.IDAsdos2 != nil {
		if err := tx.First(&models.AsistenDosen{}, "id = ?", *input.IDAsdos2).Error; err != nil {
			tx.Rollback()
			return SessionResponse{}, fmt.Errorf("assistant lecturer 2 with ID %s not found", *input.IDAsdos2)
		}
	}

	// Clash detection scoped to semester (Rule 3)
	if err := checkScheduleClash(input, startTime, endTime, ""); err != nil {
		tx.Rollback()
		return SessionResponse{}, err
	}

	// Persist new session
	session := models.JadwalUtama{
		IDKelas:       input.IDKelas,
		IDMk:          input.IDMk,
		IDSemester:    input.IDSemester,
		IDRuangan:     input.IDRuangan,
		IDAsdos1:      input.IDAsdos1,
		IDAsdos2:      input.IDAsdos2,
		IDDosen:       input.IDDosen,
		KelasMulai:    startTime,
		KelasBerakhir: endTime,
	}
	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("failed to create session: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return SessionResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return loadSessionRelations(&session)
}

// ─────────────────────────────────────────────
// Service: GetAllSessions
// ─────────────────────────────────────────────

func GetAllSessions(page, limit int, semesterID string) ([]SessionResponse, int64, error) {
	var sessions []models.JadwalUtama
	var total int64

	offset := (page - 1) * limit
	query := config.DB.Model(&models.JadwalUtama{})

	if semesterID != "" {
		query = query.Where("id_semester = ?", semesterID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	responses := make([]SessionResponse, 0, len(sessions))
	for i := range sessions {
		resp, err := loadSessionRelations(&sessions[i])
		if err != nil {
			return nil, 0, err
		}
		responses = append(responses, resp)
	}

	return responses, total, nil
}

// ─────────────────────────────────────────────
// Service: GetSessionByID
// ─────────────────────────────────────────────

func GetSessionByID(id string) (SessionResponse, error) {
	var session models.JadwalUtama
	if err := config.DB.First(&session, "id = ?", id).Error; err != nil {
		return SessionResponse{}, fmt.Errorf("session with ID %s not found", id)
	}

	return loadSessionRelations(&session)
}

// ─────────────────────────────────────────────
// Service: UpdateSession
// ─────────────────────────────────────────────

func UpdateSession(id string, input *SessionInput) (SessionResponse, error) {
	normalizeSessionInput(input)

	if err := validateInstructorXOR(input); err != nil {
		return SessionResponse{}, err
	}

	startTime, endTime, err := TranslateSchedule(input.DayOption, input.SlotOption)
	if err != nil {
		return SessionResponse{}, err
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Ensure session exists
	var session models.JadwalUtama
	if err := tx.First(&session, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("session with ID %s not found", id)
	}

	// Validate mandatory entity existence
	if err := tx.First(&models.Kelas{}, "id = ?", input.IDKelas).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("class with ID %s not found", input.IDKelas)
	}
	if err := tx.First(&models.MataKuliah{}, "id = ?", input.IDMk).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("course with ID %s not found", input.IDMk)
	}
	if err := tx.First(&models.Semester{}, "id = ?", input.IDSemester).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("semester with ID %s not found", input.IDSemester)
	}
	if err := tx.First(&models.Ruangan{}, "id = ?", input.IDRuangan).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("room with ID %s not found", input.IDRuangan)
	}

	// Validate optional entity existence
	if input.IDDosen != nil {
		if err := tx.First(&models.Dosen{}, "id = ?", *input.IDDosen).Error; err != nil {
			tx.Rollback()
			return SessionResponse{}, fmt.Errorf("lecturer with ID %s not found", *input.IDDosen)
		}
	}
	if input.IDAsdos1 != nil {
		if err := tx.First(&models.AsistenDosen{}, "id = ?", *input.IDAsdos1).Error; err != nil {
			tx.Rollback()
			return SessionResponse{}, fmt.Errorf("assistant lecturer 1 with ID %s not found", *input.IDAsdos1)
		}
	}
	if input.IDAsdos2 != nil {
		if err := tx.First(&models.AsistenDosen{}, "id = ?", *input.IDAsdos2).Error; err != nil {
			tx.Rollback()
			return SessionResponse{}, fmt.Errorf("assistant lecturer 2 with ID %s not found", *input.IDAsdos2)
		}
	}

	// Clash detection scoped to semester, excluding this session (Rule 3)
	if err := checkScheduleClash(input, startTime, endTime, id); err != nil {
		tx.Rollback()
		return SessionResponse{}, err
	}

	updates := map[string]interface{}{
		"id_kelas":       input.IDKelas,
		"id_mk":          input.IDMk,
		"id_semester":    input.IDSemester,
		"id_ruangan":     input.IDRuangan,
		"id_dosen":       input.IDDosen,
		"id_asdos1":      input.IDAsdos1,
		"id_asdos2":      input.IDAsdos2,
		"kelas_mulai":    startTime,
		"kelas_berakhir": endTime,
	}
	if err := tx.Model(&session).Updates(updates).Error; err != nil {
		tx.Rollback()
		return SessionResponse{}, fmt.Errorf("failed to update session: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return SessionResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload updated record
	if err := config.DB.First(&session, "id = ?", id).Error; err != nil {
		return SessionResponse{}, fmt.Errorf("failed to reload session after update: %w", err)
	}

	return loadSessionRelations(&session)
}

// ─────────────────────────────────────────────
// Service: DeleteSession
// ─────────────────────────────────────────────

func DeleteSession(id string) error {
	var session models.JadwalUtama
	if err := config.DB.First(&session, "id = ?", id).Error; err != nil {
		return fmt.Errorf("session with ID %s not found", id)
	}
	return config.DB.Delete(&session).Error
}

func GetDailyAssistantSessions(dateStr string, assistantID string) ([]DailySessionResponse, error) {
	db := config.DB
	var results []DailySessionResponse

	// 1. Parse Target Date
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("[ScheduleService] Failed to parse date %s: %v", dateStr, err)
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	targetDow := int(targetDate.Weekday()) // 0=Sunday, 1=Monday, ..., 6=Saturday

	// 2. Fetch Regular Schedules (Based on day of week)
	var regularSchedules []models.JadwalUtama
	// Using EXTRACT(DOW) for PostgreSQL compatibility
	if err := db.Where("(id_asdos1 = ? OR id_asdos2 = ?) AND EXTRACT(DOW FROM kelas_mulai) = ?", assistantID, assistantID, targetDow).Find(&regularSchedules).Error; err != nil {
		log.Printf("[ScheduleService] Error fetching regular schedules for assistant %s: %v", assistantID, err)
		return nil, fmt.Errorf("failed to fetch regular schedules: %w", err)
	}

	// 3. BLACKLIST: Find sessions cancelled/moved on this specific date
	var blacklistSubs []models.SubstituteSession
	if err := db.Where("status = ? AND original_date = ?", models.StatusVerified, dateStr).Find(&blacklistSubs).Error; err != nil {
		log.Printf("[ScheduleService] Error fetching blacklist for date %s: %v", dateStr, err)
		return nil, fmt.Errorf("failed to check cancelled classes: %w", err)
	}
	blacklist := make(map[string]bool)
	for _, sub := range blacklistSubs {
		blacklist[sub.IDSession] = true
	}

	// 4. Process Regular Schedules into Results (Skip blacklisted)
	for i := range regularSchedules {
		if blacklist[regularSchedules[i].ID] {
			continue // SKIP! This session is cancelled/moved today
		}

		resp, err := loadSessionRelations(&regularSchedules[i])
		if err != nil {
			log.Printf("[ScheduleService] Failed to load relations for session %s: %v", regularSchedules[i].ID, err)
			continue
		}

		// Update Waktu format: "YYYY-MM-DD, HH:MM - HH:MM"
		resp.Waktu = fmt.Sprintf("%s, %s - %s",
			dateStr,
			regularSchedules[i].KelasMulai.In(wib).Format("15:04"),
			regularSchedules[i].KelasBerakhir.In(wib).Format("15:04"),
		)

		results = append(results, DailySessionResponse{
			SessionResponse: resp,
			TipeJadwal:      "REGULAR",
		})
	}

	// 5. SUBSTITUTE SESSIONS: Fetch using optimized GORM JOINs
	var substituteSessions []models.SubstituteSession
	err = db.Preload("Session").
		Preload("Dosen").
		Preload("Asdos1").Preload("Asdos1.User").
		Preload("Asdos2").Preload("Asdos2.User").
		Joins("JOIN jadwal_utamas ju ON ju.id = substitute_sessions.id_session").
		Where("substitute_sessions.status = ? AND substitute_sessions.substitute_date = ?", models.StatusVerified, dateStr).
		Where("(substitute_sessions.id_asdos1 = ? OR substitute_sessions.id_asdos2 = ?) OR ((substitute_sessions.id_dosen IS NULL AND substitute_sessions.id_asdos1 IS NULL AND substitute_sessions.id_asdos2 IS NULL) AND (ju.id_asdos1 = ? OR ju.id_asdos2 = ?))",
			assistantID, assistantID, assistantID, assistantID).
		Find(&substituteSessions).Error

	if err != nil {
		log.Printf("[ScheduleService] Error fetching substitute sessions for assistant %s on %s: %v", assistantID, dateStr, err)
		return nil, fmt.Errorf("failed to fetch substitute sessions: %w", err)
	}

	// 6. Process Substitute Sessions into Results
	for i := range substituteSessions {
		sub := &substituteSessions[i]

		if sub.Session == nil {
			log.Printf("[ScheduleService] Substitute session %s has no linked main session", sub.ID)
			continue
		}

		resp, err := loadSessionRelations(sub.Session)
		if err != nil {
			log.Printf("[ScheduleService] Failed to load main relations for sub session %s: %v", sub.ID, err)
			continue
		}

		// Override Pengajar if there is a substitute formation
		resp.Pengajar = FormatInstructor(sub.Session, sub)

		// Override Waktu: "YYYY-MM-DD, HH:MM - HH:MM"
		resp.Waktu = fmt.Sprintf("%s, %s - %s",
			dateStr,
			sub.KelasMulai.In(wib).Format("15:04"),
			sub.KelasBerakhir.In(wib).Format("15:04"),
		)

		log.Printf("[ScheduleService] Including substitute session: %s (%s) for assistant %s",
			resp.MataKuliah, resp.NamaKelas, resp.Pengajar)

		results = append(results, DailySessionResponse{
			SessionResponse: resp,
			TipeJadwal:      "PENGGANTI",
		})
	}

	return results, nil
}
