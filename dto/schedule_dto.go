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
