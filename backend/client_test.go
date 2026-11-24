package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

// roundTripFunc — пользовательский тип для подмены HTTP-транспорта в тестах.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// exchangeWithoutAuth - Exchange без авторизации (для тестов без токена).
type exchangeWithoutAuth struct {
	bodyData any
	path     string
	response any
}

func (e exchangeWithoutAuth) Auth() bool      { return false }
func (e exchangeWithoutAuth) Path() string    { return e.path }
func (e exchangeWithoutAuth) Body() any       { return e.bodyData }
func (e exchangeWithoutAuth) Response() any   { return e.response }

// exchangeWithAuth - Exchange с авторизацией (для тестов с токеном).
type exchangeWithAuth struct {
	bodyData any
	path     string
	response any
}

func (e exchangeWithAuth) Auth() bool      { return true }
func (e exchangeWithAuth) Path() string    { return e.path }
func (e exchangeWithAuth) Body() any       { return e.bodyData }
func (e exchangeWithAuth) Response() any   { return e.response }

// errorExchange - Exchange, возвращающий ошибку API.
type errorExchange struct {
	bodyData any
	path     string
}

func (e errorExchange) Auth() bool      { return true }
func (e errorExchange) Path() string    { return e.path }
func (e errorExchange) Body() any       { return e.bodyData }
func (e errorExchange) Response() any   { return nil }

// ErrTransport — пользовательская ошибка для тестов транспорта.
type ErrTransport string

func (e ErrTransport) Error() string {
	return string(e)
}

func TestDo_SuccessWithoutAuth(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "", false)

	// Настраиваем mock transport
	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		rec.Write([]byte(`{}`))
		return rec.Result(), nil
	}))

	resp, err := c.Do(context.Background(), exchangeWithoutAuth{
		bodyData: map[string]any{"key": "value"},
		path:     "/api/v1/test",
		response: struct{}{},
	})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response, got nil")
	}
}

func TestDo_SuccessWithAuth(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "test-token-123", false)

	// Устанавливаем действительный токен
	c.mu.Lock()
	c.token = "test-token-123"
	c.tokenExpireIn = time.Now().Add(24 * time.Hour)
	c.mu.Unlock()

	// Настраиваем mock transport
	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// Проверяем заголовок авторизации
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-123" {
			t.Errorf("Expected Authorization 'Bearer test-token-123', got '%s'", auth)
		}
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		rec.Write([]byte(`{}`))
		return rec.Result(), nil
	}))

	resp, err := c.Do(context.Background(), exchangeWithAuth{
		bodyData: map[string]any{"key": "value"},
		path:     "/api/v1/test",
		response: struct{}{},
	})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response, got nil")
	}
}

func TestDo_APIError(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "test-token-123", false)

	c.mu.Lock()
	c.token = "test-token-123"
	c.tokenExpireIn = time.Now().Add(24 * time.Hour)
	c.mu.Unlock()

	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusNotFound)
		rec.Write([]byte(`{"code":"receipt.fiscaldata.not.found.dr","message":"Not found"}`))
		return rec.Result(), nil
	}))

	req := errorExchange{
		bodyData: map[string]any{"key": "value"},
		path:     "/api/v1/test",
	}

	_, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected API error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected *APIError, got %T", err)
	}
	if apiErr.Code != "receipt.fiscaldata.not.found.dr" {
		t.Fatalf("Expected code 'receipt.fiscaldata.not.found.dr', got '%s'", apiErr.Code)
	}
	if apiErr.Message != "Not found" {
		t.Fatalf("Expected message 'Not found', got '%s'", apiErr.Message)
	}
}

func TestDo_TransportError(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "", false)

	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, ErrTransport("network error")
	}))

	req := exchangeWithoutAuth{
		bodyData: map[string]any{"key": "value"},
		path:     "/api/v1/test",
		response: struct{}{},
	}

	_, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected transport error, got nil")
	}
	var transportErr *url.Error
	if !errors.As(err, &transportErr) {
		t.Fatalf("Expected *url.Error, got %T", err)
	}
	if transportErr.Err.Error() != "network error" {
		t.Fatalf("Expected 'network error', got '%v'", err)
	}
}

func TestDo_InvalidJSON(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "", false)

	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		rec.Write([]byte("not valid json"))
		return rec.Result(), nil
	}))

	req := exchangeWithoutAuth{
		bodyData: map[string]any{"key": "value"},
		path:     "/api/v1/test",
		response: struct{}{},
	}

	_, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected JSON decode error, got nil")
	}
}

func TestDo_AutoTokenRefresh(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "old-refresh-token", false)

	// Устанавливаем истёкший токен (более 5 минут назад)
	c.mu.Lock()
	c.token = "old-token"
	c.tokenExpireIn = time.Now().Add(-10 * time.Minute)
	c.mu.Unlock()

	var mu sync.Mutex
	refreshCalled := false

	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()

		if r.URL.Path == "/api/v1/auth/token" && !refreshCalled {
			// Refresh-запрос — возвращаем новый токен
			refreshCalled = true
			rec := httptest.NewRecorder()
			rec.WriteHeader(http.StatusOK)
			rec.Write([]byte(`{"token":"new-token-456","tokenExpireIn":"2026-01-15T13:00:00.000Z","refreshToken":"new-refresh-token"}`))
			return rec.Result(), nil
		}

		// Основной запрос с новым токеном
		if r.URL.Path == "/api/v1/test" {
			rec := httptest.NewRecorder()
			rec.WriteHeader(http.StatusOK)
			rec.Write([]byte(`{}`))
			return rec.Result(), nil
		}

		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusNotFound)
		return rec.Result(), nil
	}))

	ctx := context.Background()

	// Первый вызов - должен обновить токен и выполнить запрос
	resp, err := c.Do(ctx, exchangeWithAuth{
		bodyData: map[string]any{"key": "value"},
		path:     "/api/v1/test",
		response: struct{}{},
	})
	if err != nil {
		t.Fatalf("First Do failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	// Проверяем, что токен обновился
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()
	if token != "new-token-456" {
		t.Fatalf("Expected token 'new-token-456', got '%s'", token)
	}

	// Проверяем, что refresh вызывался
	if !refreshCalled {
		t.Fatal("Expected refresh to be called")
	}
}

func TestDo_ConcurrentAuthCalls(t *testing.T) {
	c := NewClient("http://localhost:8080", DeviceInfo{}, "concurrent-token", false)

	c.mu.Lock()
	c.token = "concurrent-token"
	c.tokenExpireIn = time.Now().Add(24 * time.Hour)
	c.mu.Unlock()

	var mu sync.Mutex
	callCount := 0

	c.SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		callCount++
		mu.Unlock()

		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		rec.Write([]byte(`{}`))
		return rec.Result(), nil
	}))

	var wg sync.WaitGroup
	numCalls := 10

	// Запускаем 10 параллельных вызовов с авторизацией
	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Do(context.Background(), exchangeWithAuth{
				bodyData: map[string]any{"key": "value"},
				path:     "/api/v1/test",
				response: struct{}{},
			})
			if err != nil {
				t.Logf("Call error (non-fatal): %v", err)
			}
		}()
	}

	wg.Wait()

	// Все вызовы должны пройти успешно
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()
	if token == "" {
		t.Fatal("Token should not be empty after concurrent calls")
	}
}
