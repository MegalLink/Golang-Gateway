package utils

import (
	"time"
)

// GetCurrentDate retorna la fecha y hora actual en el formato especificado para una zona horaria dada
func GetCurrentDate(timezone string) string {
	return time.Now().In(time.FixedZone(timezone, 0)).Format("2006-01-02 15:04:05")
}

// GetTimeField retorna la hora actual en formato "150405" (HHmmss) para el campo F12
func GetTimeField(timezone string) string {
	return time.Now().In(time.FixedZone(timezone, 0)).Format("150405")
}

// GetDateField retorna la fecha actual en formato "0102" (MMDD) para el campo F13
func GetDateField(timezone string) string {
	return time.Now().In(time.FixedZone(timezone, 0)).Format("0102")
}

// GetFullDateField retorna la fecha actual en formato "20060102" (YYYYMMDD)
func GetFullDateField(timezone string) string {
	return time.Now().In(time.FixedZone(timezone, 0)).Format("20060102")
}
