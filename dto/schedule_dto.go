package dto

type ScheduleRequest struct {
	ClassCode   string `json:"class_code"`
	ClassName   string `json:"class_name"`
	SubjectCode string `json:"subject_code"`
	TeacherNIK  string `json:"teacher_nik"`
	TeacherName string `json:"teacher_name"`
	Date        string `json:"date"`
	JamKe       int    `json:"jam_ke"`
	TimeStart   string `json:"time_start"`
	TimeEnd     string `json:"time_end"`
}

type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type PeriodResponse struct {
	StartDate string `json:"start_date" example:"2025-02-01"`
	EndDate   string `json:"end_date" example:"2025-02-28"`
}

type FoundationRecapClassDetail struct {
	ClassCode string `json:"class_code" example:"XA01"`
	ClassName string `json:"class_name" example:"X-A"`
	JumlahJP  int64  `json:"jumlah_jp" example:"15"`
}

type FoundationRecapTeacher struct {
	TeacherNIK  string                       `json:"teacher_nik" example:"20222029"`
	TeacherName string                       `json:"teacher_name" example:"Najdin Aqmarina, S.Pd."`
	TotalJP     int64                        `json:"total_jp" example:"40"`
	TotalKelas  int                          `json:"total_kelas" example:"3"`
	Detail      []FoundationRecapClassDetail `json:"detail"`
}

type FoundationRecapResponse struct {
	Periode       PeriodResponse           `json:"periode"`
	TotalPengajar int                      `json:"total_pengajar" example:"3"`
	Rekap         []FoundationRecapTeacher `json:"rekap"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Unauthorized"`
}

type ErrorWithDetailsResponse struct {
	Error   string   `json:"error" example:"Validation failed"`
	Details []string `json:"details"`
}
