package auth

import (
	"context"
	"encoding/json"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type UserPreferences struct {
	Language         string `json:"language"`
	ContentLanguage  string `json:"content_language"`
	ItemsPerPage     int    `json:"items_per_page"`
	PopularTagsLimit int    `json:"popular_tags_limit"`
	HideMature       bool   `json:"hide_mature"`
}

type AppSettings struct {
	DefaultLanguage      string  `json:"default_language"`
	DefaultContentLang   string  `json:"default_content_language"`
	ItemsPerPage         int     `json:"items_per_page"`
	PopularTagsLimit     int     `json:"popular_tags_limit"`
	SiteName             string  `json:"site_name"`
	MaintenanceMode      bool    `json:"maintenance_mode"`
	RegistrationsOpen    bool    `json:"registrations_open"`
	MaxUploadSizeMB      int     `json:"max_upload_size_mb"`
	ImageServingMode     string  `json:"image_serving_mode"`
	ImgproxyBaseURL      string  `json:"imgproxy_base_url"`
	ImgproxyKey          string  `json:"imgproxy_key"`
	ImgproxySalt         string  `json:"imgproxy_salt"`
	UploadMode           string  `json:"upload_mode"`
	RequireEmailVerify   bool    `json:"require_email_verify"`
	RateLimit            int     `json:"rate_limit"`
	S3PresignedTTLMin    int     `json:"s3_presigned_ttl_min"`
	CFPresignedTTLMin    int     `json:"cf_presigned_ttl_min"`
	Boost1Downloads      int     `json:"boost_1_downloads"`
	Boost1Price          float64 `json:"boost_1_price"`
	Boost2Downloads      int     `json:"boost_2_downloads"`
	Boost2Price          float64 `json:"boost_2_price"`
	Boost3Downloads      int     `json:"boost_3_downloads"`
	Boost3Price          float64 `json:"boost_3_price"`
	ContactEmail         string `json:"contact_email"`
	HideMatureDefault    bool   `json:"hide_mature_default"`
	ForbidMatureForFree  bool   `json:"forbid_mature_for_free"`
	EnableComments       bool   `json:"enable_comments"`
	DefaultMetaDescription string `json:"default_meta_description"`

	// AI moderation (ADR 0018)
	AIModerationEnabled      bool    `json:"ai_moderation_enabled"`
	AIModel                  string  `json:"ai_model"`
	AIEndpoint               string  `json:"ai_endpoint"`
	AIPrompt                 string  `json:"ai_prompt"`
	AIAutoApproveThreshold   float64 `json:"ai_auto_approve_threshold"`
	AIAutoRejectThreshold    float64 `json:"ai_auto_reject_threshold"`

	// Billing hygiene: downgrade subscriptions stuck in "waiting for payment".
	WaitingPayJobEnabled   bool `json:"waiting_pay_job_enabled"`
	WaitingPayExpiryHours  int  `json:"waiting_pay_expiry_hours"`

	// Download delivery: archives above this size are served via a presigned
	// redirect instead of being streamed through the backend.
	DownloadStreamThresholdMB int `json:"download_stream_threshold_mb"`

	// Upload page: above this page count the manual tab switches from image
	// thumbnails to a compact file list.
	PagePreviewThreshold int `json:"page_preview_threshold"`

	// Upload: archives larger than this are split into multiple parts.
	UploadPartSizeMB int `json:"upload_part_size_mb"`

	// Upload: number of parts uploaded concurrently when splitting an archive.
	UploadConcurrency int `json:"upload_concurrency"`
}

var defaultPreferences = UserPreferences{
	Language:         "en",
	ContentLanguage:  "en",
	ItemsPerPage:     12,
	PopularTagsLimit: 20,
	HideMature:       false,
}

//encore:api auth method=GET path=/me/preferences
func GetPreferences(ctx context.Context) (*UserPreferences, error) {
	data := auth.Data().(*AuthData)

	// Try user-specific preferences first
	var prefsJSON []byte
	err := db.QueryRow(ctx, `SELECT preferences FROM users WHERE id = $1`, data.UserID).Scan(&prefsJSON)
	if err != nil || len(prefsJSON) <= 2 {
		// No user prefs → return global defaults
		return getGlobalDefaults(ctx)
	}

	var prefs UserPreferences
	if err := json.Unmarshal(prefsJSON, &prefs); err != nil {
		return getGlobalDefaults(ctx)
	}
	return &prefs, nil
}

//encore:api auth method=PATCH path=/me/preferences
func SavePreferences(ctx context.Context, p *UserPreferences) (*UserPreferences, error) {
	data := auth.Data().(*AuthData)

	b, _ := json.Marshal(p)
	_, err := db.Exec(ctx, `UPDATE users SET preferences = $1 WHERE id = $2`, string(b), data.UserID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetUserPreferences exposes a user's preferences (merged with global defaults)
// to other services that cannot read the auth database directly (ADR 0016).
//encore:api private method=GET path=/auth/user-preferences/:userID
func GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	var prefsJSON []byte
	err := db.QueryRow(ctx, `SELECT preferences FROM users WHERE id = $1`, userID).Scan(&prefsJSON)
	if err != nil || len(prefsJSON) <= 2 {
		return getGlobalDefaults(ctx)
	}

	var prefs UserPreferences
	if err := json.Unmarshal(prefsJSON, &prefs); err != nil {
		return getGlobalDefaults(ctx)
	}
	return &prefs, nil
}

func getGlobalDefaults(ctx context.Context) (*UserPreferences, error) {
	var raw []byte
	err := db.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = 'defaults'`).Scan(&raw)
	if err != nil || len(raw) == 0 {
		result := defaultPreferences
		return &result, nil
	}

	var settings AppSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		result := defaultPreferences
		return &result, nil
	}

	return &UserPreferences{
		Language:         settings.DefaultLanguage,
		ContentLanguage:  settings.DefaultContentLang,
		ItemsPerPage:     settings.ItemsPerPage,
		PopularTagsLimit: settings.PopularTagsLimit,
		HideMature:       false,
	}, nil
}

// ----- Admin Settings -----

// loadSettings returns the merged global settings (defaults overlaid with the
// stored blob), falling back to defaults when none are stored.
func loadSettings(ctx context.Context) *AppSettings {
	settings := *defaultAppSettings()
	var raw []byte
	if err := db.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = 'defaults'`).Scan(&raw); err == nil && len(raw) > 0 {
		json.Unmarshal(raw, &settings)
	}
	return &settings
}

// GetAppConfig exposes the merged global settings to other services that are
// not allowed to read the auth database directly (ADR 0016).
//encore:api private method=GET path=/auth/app-config
func GetAppConfig(ctx context.Context) (*AppSettings, error) {
	return loadSettings(ctx), nil
}

// SiteConfig is the public, non-sensitive subset of settings used for SEO,
// metadata and footer content.
type SiteConfig struct {
	SiteName               string `json:"site_name"`
	ContactEmail           string `json:"contact_email"`
	DefaultMetaDescription string `json:"default_meta_description"`
	DefaultLanguage        string `json:"default_language"`
	MaintenanceMode        bool   `json:"maintenance_mode"`
	UploadMode             string `json:"upload_mode"`
	PagePreviewThreshold   int    `json:"page_preview_threshold"`
	UploadPartSizeMB       int    `json:"upload_part_size_mb"`
	UploadConcurrency      int    `json:"upload_concurrency"`
}

//encore:api public method=GET path=/site-config
func GetSiteConfig(ctx context.Context) (*SiteConfig, error) {
	s := loadSettings(ctx)
	return &SiteConfig{
		SiteName:               s.SiteName,
		ContactEmail:           s.ContactEmail,
		DefaultMetaDescription: s.DefaultMetaDescription,
		DefaultLanguage:        s.DefaultLanguage,
		MaintenanceMode:        s.MaintenanceMode,
		UploadMode:             s.UploadMode,
		PagePreviewThreshold:   s.PagePreviewThreshold,
		UploadPartSizeMB:       s.UploadPartSizeMB,
		UploadConcurrency:      s.UploadConcurrency,
	}, nil
}

//encore:api auth method=GET path=/admin/settings
func GetAdminSettings(ctx context.Context) (*AppSettings, error) {
	ad := auth.Data().(*AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	return loadSettings(ctx), nil
}

//encore:api auth method=PATCH path=/admin/settings
func SaveAdminSettings(ctx context.Context, p *AppSettings) (*AppSettings, error) {
	ad := auth.Data().(*AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	b, _ := json.Marshal(p)
	_, err := db.Exec(ctx, `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES ('defaults', $1, now())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = now()
	`, string(b))
	if err != nil {
		return nil, err
	}

	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'update_settings', 'settings', 'global', 'settings updated')`,
		ad.UserID)

	return p, nil
}

func defaultAppSettings() *AppSettings {
	return &AppSettings{
		DefaultLanguage:    "en",
		DefaultContentLang: "en",
		ItemsPerPage:       12,
		PopularTagsLimit:   20,
		SiteName:           "Comics Galore",
		RegistrationsOpen:  true,
		MaxUploadSizeMB:    3000,
		ImageServingMode:   "direct",
		UploadMode:         "backend",
		RateLimit:           60,
		S3PresignedTTLMin:   15,
		CFPresignedTTLMin:   15,
		Boost1Downloads:     10,
		Boost1Price:         5,
		Boost2Downloads:     25,
		Boost2Price:         10,
		Boost3Downloads:         60,
		Boost3Price:             20,
		ContactEmail:         "",
		HideMatureDefault:    false,
		ForbidMatureForFree:  false,
		EnableComments:       false,
		DefaultMetaDescription: "",

		AIModerationEnabled:    false,
		AIModel:                "gpt-4o-mini",
		AIEndpoint:             "https://api.openai.com/v1/chat/completions",
		AIPrompt:               "You moderate user-generated content on a comics platform. Reply with only JSON: {\"decision\":\"approved|rejected|uncertain\",\"confidence\":0.0,\"reason\":\"...\"}.",
		AIAutoApproveThreshold: 0.85,
		AIAutoRejectThreshold:  0.15,

		WaitingPayJobEnabled:  true,
		WaitingPayExpiryHours: 24,

		DownloadStreamThresholdMB: 10,
		PagePreviewThreshold:      20,
		UploadPartSizeMB:          100,
		UploadConcurrency:         4,
	}
}
