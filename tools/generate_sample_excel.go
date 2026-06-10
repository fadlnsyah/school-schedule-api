package main

import (
	"log"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	if err := os.MkdirAll("examples", 0755); err != nil {
		log.Fatal(err)
	}

	rows := [][]any{
		{"class_code", "class_name", "subject_code", "teacher_nik", "teacher_name", "date", "jam_ke", "time_start", "time_end"},
		{"XA01", "X-A", "BIO", "20222028", "Amalia Putri, S.Pd.", "2025-02-10", 1, "07:00:00", "07:40:00"},
		{"XA01", "X-A", "CHEM", "20222029", "Najdin Aqmarina, S.Pd.", "2025-02-10", 2, "08:40:00", "09:20:00"},
		{"XB01", "X-B", "MATH", "20222030", "Rizky Maulana, S.Pd.", "2025-02-10", 1, "07:00:00", "07:40:00"},
		{"XA01", "X-A", "MATH", "20222030", "Rizky Maulana, S.Pd.", "2025-02-11", 1, "07:00:00", "07:40:00"},
		{"XB01", "X-B", "BIO", "20222028", "Amalia Putri, S.Pd.", "2025-02-11", 2, "08:40:00", "09:20:00"},
		{"XA01", "X-A", "CHEM", "20222029", "Najdin Aqmarina, S.Pd.", "2025-02-12", 3, "09:20:00", "10:00:00"},
		{"XB01", "X-B", "CHEM", "20222029", "Najdin Aqmarina, S.Pd.", "2025-02-13", 2, "08:40:00", "09:20:00"},
		{"XA01", "X-A", "BIO", "20222028", "Amalia Putri, S.Pd.", "2025-02-14", 1, "07:00:00", "07:40:00"},
	}

	f := excelize.NewFile()
	sheet := "Schedules"
	f.SetSheetName("Sheet1", sheet)

	for r, row := range rows {
		for c, value := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}

	_ = f.SetColWidth(sheet, "A", "I", 18)
	if err := f.SaveAs("examples/sample-schedules.xlsx"); err != nil {
		log.Fatal(err)
	}
}
