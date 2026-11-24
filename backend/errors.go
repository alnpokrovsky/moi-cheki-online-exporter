package main

import (
	"errors"
	"fmt"
)

// APIError — структурированная ошибка API ФНС с кодом и сообщением.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		if e.Message != "" {
			return fmt.Sprintf("%s (%s)", e.Code, e.Message)
		}
		return e.Code
	}
	if e.Message != "" {
		return e.Message
	}
	return ""
}

// IsDataNotFound возвращает true, если ошибка API указывает на отсутствие фискальных данных чека.
func IsDataNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == "receipt.fiscaldata.not.found.dr"
	}
	return false
}
