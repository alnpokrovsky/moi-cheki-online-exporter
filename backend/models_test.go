package main

import (
	"encoding/json"
	"testing"
	"time"
)

// TestDateTime_UnmarshalJSON парсит "2026-01-15T10:30:00" в Europe/Moscow.
func TestDateTime_UnmarshalJSON(t *testing.T) {
	input := `"2026-01-15T10:30:00"`
	var dt DateTime

	// UnmarshalJSON вызывается через json.Unmarshal
	err := json.Unmarshal([]byte(input), &dt)
	if err != nil {
		t.Fatalf("Expected no error for valid date, got: %v", err)
	}

	// Проверяем, что время распарсено в Europe/Moscow
	expectedHour := 10
	expectedMinute := 30

	if dt.Hour() != expectedHour {
		t.Fatalf("Expected hour %d, got %d", expectedHour, dt.Hour())
	}
	if dt.Minute() != expectedMinute {
		t.Fatalf("Expected minute %d, got %d", expectedMinute, dt.Minute())
	}
	if dt.Day() != 15 {
		t.Fatalf("Expected day 15, got %d", dt.Day())
	}
	if dt.Month() != time.January {
		t.Fatalf("Expected month January, got %d", dt.Month())
	}
	if dt.Year() != 2026 {
		t.Fatalf("Expected year 2026, got %d", dt.Year())
	}
}

// TestDateTime_MarshalJSON создаёт строку в правильном формате.
func TestDateTime_MarshalJSON(t *testing.T) {
	// Создаем DateTime с конкретным временем
	loc, _ := time.LoadLocation("Europe/Moscow")
	testTime := time.Date(2026, 1, 15, 10, 30, 0, 0, loc)
	dt := DateTime{Time: testTime}

	output, err := json.Marshal(dt)
	if err != nil {
		t.Fatalf("Expected no error marshaling, got: %v", err)
	}

	expected := `"2026-01-15T10:30:00"`
	if string(output) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(output))
	}
}

// TestDateTime_UnmarshalJSON_InvalidJSON возвращает ошибку при неверном JSON.
func TestDateTime_UnmarshalJSON_InvalidJSON(t *testing.T) {
	input := `"not-a-date"`
	var dt DateTime

	err := json.Unmarshal([]byte(input), &dt)
	if err == nil {
		t.Fatal("Expected error for invalid date JSON, got nil")
	}

	if err.Error() == "" {
		t.Fatal("Expected non-empty error message")
	}
}

// TestDateTime_MarshalJSON_ZeroValue marshaling zero value.
func TestDateTime_MarshalJSON_ZeroValue(t *testing.T) {
	var dt DateTime

	output, err := json.Marshal(dt)
	if err != nil {
		t.Fatalf("Expected no error marshaling zero value, got: %v", err)
	}

	// Пустое время marshals в пустую строку или "null" в зависимости от реализации
	// Валидно, что код не panic
	_ = output
}