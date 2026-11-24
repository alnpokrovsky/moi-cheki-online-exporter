package main

type DeviceInfo struct {
	SourceDeviceId string `json:"sourceDeviceId"`
	SourceType     string `json:"sourceType"`
	AppVersion     string `json:"appVersion"`
	MetaDetails    struct {
		UserAgent string `json:"userAgent"`
	} `json:"metaDetails"`
}

type RefreshTokenRequest struct {
	RefreshToken string     `json:"refreshToken"`
	DeviceInfo   DeviceInfo `json:"deviceInfo"`
}

type TokenResponse struct {
	RefreshToken  string `json:"refreshToken"`
	Token         string `json:"token"`
	TokenExpireIn string `json:"tokenExpireIn"`
}
