package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
)

type DateOnly pgtype.Date

type TimeOnly pgtype.Time

type Schedule struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ClassCode   string    `json:"class_code" gorm:"type:varchar(10);not null;index;index:idx_class_date,priority:1"`
	ClassName   string    `json:"class_name" gorm:"type:varchar(10);not null"`
	SubjectCode string    `json:"subject_code" gorm:"type:varchar(10);not null"`
	TeacherNIK  string    `json:"teacher_nik" gorm:"type:varchar(20);not null;index;index:idx_teacher_date,priority:1"`
	TeacherName string    `json:"teacher_name" gorm:"type:varchar(100);not null"`
	Date        DateOnly  `json:"date" gorm:"type:date;not null;index;index:idx_class_date,priority:2;index:idx_teacher_date,priority:2"`
	JamKe       int       `json:"jam_ke" gorm:"not null"`
	TimeStart   TimeOnly  `json:"time_start" gorm:"type:time;not null"`
	TimeEnd     TimeOnly  `json:"time_end" gorm:"type:time;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func ParseDateOnly(value string) (DateOnly, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return DateOnly{}, err
	}

	return DateOnly(pgtype.Date{Time: parsed, Valid: true}), nil
}

func ParseTimeOnly(value string) (TimeOnly, error) {
	parsed, err := time.Parse("15:04:05", value)
	if err != nil {
		return TimeOnly{}, err
	}

	microseconds := int64(parsed.Hour()*60*60+parsed.Minute()*60+parsed.Second()) * 1000000
	return TimeOnly(pgtype.Time{Microseconds: microseconds, Valid: true}), nil
}

func (d DateOnly) String() string {
	date := pgtype.Date(d)
	if !date.Valid {
		return ""
	}
	return date.Time.Format("2006-01-02")
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d DateOnly) Value() (driver.Value, error) {
	date := pgtype.Date(d)
	return date.Value()
}

func (d *DateOnly) Scan(src any) error {
	var date pgtype.Date
	if err := date.Scan(src); err != nil {
		return err
	}
	*d = DateOnly(date)
	return nil
}

func (t TimeOnly) String() string {
	value := pgtype.Time(t)
	if !value.Valid {
		return ""
	}

	totalSeconds := value.Microseconds / 1000000
	hour := totalSeconds / 3600
	minute := (totalSeconds % 3600) / 60
	second := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

func (t TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t TimeOnly) Value() (driver.Value, error) {
	value := pgtype.Time(t)
	return value.Value()
}

func (t *TimeOnly) Scan(src any) error {
	var value pgtype.Time
	if err := value.Scan(src); err != nil {
		return err
	}
	*t = TimeOnly(value)
	return nil
}
