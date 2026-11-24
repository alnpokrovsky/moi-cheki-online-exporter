package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const expireTokenOffset = 5 * time.Minute

// Exchange описывает API-запрос и связывает его с типом ответа.
type Exchange[R any] interface {
	Auth() bool      // требует ли запрос Bearer-авторизацию
	Path() string    // относительный путь API, например "/api/v1/receipt"
	Body() any       // тело запроса для POST (marshal-ируется в Do)
	Response() R     // нулевой экземпляр типа ответа для десериализации
}

// Client — HTTP-клиент для работы с API ФНС "Мои чеки онлайн".
type Client struct {
	baseURL      string
	httpClient   *http.Client
	deviceInfo   DeviceInfo
	refreshToken string
	token        string
	tokenExpireIn time.Time
	debugMode    bool
	mu           sync.RWMutex
}

// NewClient создаёт новый экземпляр Client.
func NewClient(baseURL string, deviceInfo DeviceInfo, refreshToken string, debugMode bool) *Client {
	return &Client{
		baseURL:      baseURL,
		httpClient:   &http.Client{},
		deviceInfo:   deviceInfo,
		refreshToken: refreshToken,
		debugMode:    debugMode,
	}
}

// SetTransport подменяет HTTP-транспорт клиента (используется в тестах).
func (c *Client) SetTransport(rt http.RoundTripper) {
	c.httpClient.Transport = rt
}

// Do выполняет API-запрос и десериализует ответ в тип R.
func (c *Client) Do[R any](ctx context.Context, in Exchange[R]) (*R, error) {
	var token string
	if in.Auth() {
		if err := c.ensureToken(ctx); err != nil {
			return nil, err
		}
		c.mu.RLock()
		token = c.token
		c.mu.RUnlock()
	}

	reqBody, err := json.Marshal(in.Body())
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+in.Path(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	if c.debugMode {
		log.Printf("DEBUG: Request Type: POST")
		log.Printf("DEBUG: Request URL: %s", httpReq.URL)
		log.Printf("DEBUG: Request Body: %s", string(reqBody))
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if c.debugMode {
		log.Printf("DEBUG: Response Status: %d", httpResp.StatusCode)
		log.Printf("DEBUG: Response Body: %s", string(body))
	}

	if httpResp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Code != "" {
			return nil, &apiErr
		}
		return nil, &APIError{Code: "", Message: httpResp.Status}
	}

	var out R
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ensureToken проверяет, не истёк ли access-токен с запасом 5 минут.
// Если истёк — выполняетrefresh через refresh-токен.
func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.RLock()
	expired := c.token == "" || time.Now().After(c.tokenExpireIn.Add(-expireTokenOffset))
	c.mu.RUnlock()

	if !expired {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check после захвата записи
	if c.token != "" && !time.Now().After(c.tokenExpireIn.Add(-expireTokenOffset)) {
		return nil
	}

	return c.refreshAccessToken(ctx)
}

// refreshAccessToken обновляет access-токен через refresh-токен (POST /api/v1/auth/token).
func (c *Client) refreshAccessToken(ctx context.Context) error {
	reqBody := RefreshTokenRequest{
		RefreshToken: c.refreshToken,
		DeviceInfo:   c.deviceInfo,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/token", bytes.NewReader(reqBodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if c.debugMode {
		log.Printf("DEBUG: Request Type: POST")
		log.Printf("DEBUG: Request URL: %s", httpReq.URL)
		log.Printf("DEBUG: Request Body: %s", string(reqBodyBytes))
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if c.debugMode {
		log.Printf("DEBUG: Response Status: %s", httpResp.Status)
		log.Printf("DEBUG: Response Body: %s", string(body))
	}

	if httpResp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Code != "" {
			return &apiErr
		}
		return &APIError{Code: "", Message: httpResp.Status}
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return err
	}

	c.token = tokenResponse.Token
	c.refreshToken = tokenResponse.RefreshToken

	if tokenResponse.TokenExpireIn == "" {
		c.tokenExpireIn = time.Now().Add(1 * time.Hour)
	} else {
		c.tokenExpireIn, err = time.Parse(time.RFC3339, tokenResponse.TokenExpireIn)
		if err != nil {
			return err
		}
	}

	if c.debugMode {
		log.Printf("DEBUG: Successfully obtained access token: %s", c.token)
		log.Printf("DEBUG: Access token expires at: %s", c.tokenExpireIn.Format(time.RFC3339))
	}

	return nil
}

// receiptsExchange реализует Exchange для получения списка чеков.
type receiptsExchange struct {
	DateFrom string
	DateTo   string
}

func (e receiptsExchange) Auth() bool      { return true }
func (e receiptsExchange) Path() string    { return "/api/v1/receipt" }
func (e receiptsExchange) Body() any {
	return map[string]any{
		"limit":    10000,
		"offset":   0,
		"dateFrom": e.DateFrom,
		"dateTo":   e.DateTo,
		"orderBy":  "CREATED_DATE:ASC",
	}
}
func (e receiptsExchange) Response() ReceiptResponse { return ReceiptResponse{} }

// fiscalDataExchange реализует Exchange для получения фискальных данных чека.
type fiscalDataExchange struct {
	Key string
}

func (e fiscalDataExchange) Auth() bool      { return true }
func (e fiscalDataExchange) Path() string    { return "/api/v1/receipt/fiscal_data" }
func (e fiscalDataExchange) Body() any       { return map[string]any{"key": e.Key} }
func (e fiscalDataExchange) Response() FiscalDataResponse { return FiscalDataResponse{} }

// tokenExchange реализует Exchange для обновления access-токена.
type tokenExchange struct {
	RefreshToken string     `json:"refreshToken"`
	DeviceInfo   DeviceInfo `json:"deviceInfo"`
}

func (e tokenExchange) Auth() bool           { return false }
func (e tokenExchange) Path() string         { return "/api/v1/auth/token" }
func (e tokenExchange) Body() any            { return e }
func (e tokenExchange) Response() TokenResponse { return TokenResponse{} }
