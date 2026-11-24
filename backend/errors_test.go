package main

import (
	"errors"
	"testing"
)

// TestAPIError_Error форматирует строку корректно.
func TestAPIError_Error(t *testing.T) {
	// С кодом и сообщением
	apiErr := &APIError{Code: "test-code", Message: "test-message"}
	expected := "test-code (test-message)"
	if apiErr.Error() != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, apiErr.Error())
	}

	// Только код
	apiErr2 := &APIError{Code: "only-code"}
	expected2 := "only-code"
	if apiErr2.Error() != expected2 {
		t.Fatalf("Expected '%s', got '%s'", expected2, apiErr2.Error())
	}

	// Только сообщение
	apiErr3 := &APIError{Message: "only-message"}
	expected3 := "only-message"
	if apiErr3.Error() != expected3 {
		t.Fatalf("Expected '%s', got '%s'", expected3, apiErr3.Error())
	}

	// Пустой
	apiErr4 := &APIError{}
	expected4 := ""
	if apiErr4.Error() != expected4 {
		t.Fatalf("Expected '%s', got '%s'", expected4, apiErr4.Error())
	}
}

// TestIsDataNotFound true для нужного кода.
func TestIsDataNotFound_True(t *testing.T) {
	apiErr := &APIError{Code: "receipt.fiscaldata.not.found.dr"}
	if !IsDataNotFound(apiErr) {
		t.Fatal("Expected IsDataNotFound to return true for 'receipt.fiscaldata.not.found.dr'")
	}
}

// TestIsDataNotFound false для других кодов.
func TestIsDataNotFound_False_OtherCodes(t *testing.T) {
	apiErr := &APIError{Code: "other.code"}
	if IsDataNotFound(apiErr) {
		t.Fatal("Expected IsDataNotFound to return false for 'other.code'")
	}
}

// TestIsDataNotFound false для не-APIError.
func TestIsDataNotFound_False_NonAPIError(t *testing.T) {
	err := errors.New("some other error")
	if IsDataNotFound(err) {
		t.Fatal("Expected IsDataNotFound to return false for non-APIError")
	}
}

// TestIsDataNotFound false для nil.
func TestIsDataNotFound_False_Nil(t *testing.T) {
	if IsDataNotFound(nil) {
		t.Fatal("Expected IsDataNotFound to return false for nil")
	}
}