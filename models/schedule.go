package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Schedule struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ClassCode   string    `json:"class_code" gorm:"type:varchar(10);not null;index;index:idx_class_date,priority:1"`
	ClassName   string    `json:"class_name" gorm:"type:varchar(10);not null"`
	SubjectCode string    `json:"subject_code" gorm:"type:varchar(10);not null"`
	TeacherNIK  string    `json:"teacher_nik" gorm:"type:varchar(20);not null;index;index:idx_teacher_date,priority:1"`
	TeacherName string    `json:"teacher_name" gorm:"type:varchar(100);not null"`
	Date        string    `json:"date" gorm:"type:date;not null;index;index:idx_class_date,priority:2;index:idx_teacher_date,priority:2"`
	JamKe       int       `json:"jam_ke" gorm:"not null"`
	TimeStart   string    `json:"time_start" gorm:"type:time;not null"`
	TimeEnd     string    `json:"time_end" gorm:"type:time;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
