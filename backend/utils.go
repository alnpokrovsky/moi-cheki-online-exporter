package main

import (
	"math"

	"github.com/skip2/go-qrcode"
)

// RoundToFloat64 округляет значение до двух знаков после запятой.
func RoundToFloat64(value float64) float64 {
	return math.Round(value*100) / 100
}

// generateQRCode генерирует QR-код для заданного текста.
func generateQRCode(text string) ([]byte, error) {
	qrCode, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	return qrCode, nil
}
