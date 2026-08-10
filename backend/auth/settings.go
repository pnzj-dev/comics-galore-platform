package auth

import (
	"context"
	"encoding/json"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type UserPreferences struct {
	Language             string `json:"language"`
	ContentLanguage      string `json:"content_language"`
	ItemsPerPage         int    `json:"items_per_page"`
	PopularTagsLimit     int    `json:"popular_tags_limit"`
	EmailFromFollowing   bool   `json:"email_from_following"`
	EmailSupportReplies  bool   `json:"email_support_replies"`
	EmailMarketing       bool   `json:"email_marketing"`
	InAppEnabled         bool   `json:"in_app_enabled"`
	HideMature           bool   `json:"hide_mature"`
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
	RequireEmailVerify   bool    `json:"require_email_verify"`
	RateLimit            int     `json:"rate_limit"`
	S3PresignedTTLMin    int     `json:"s3_presigned_ttl_min"`
	CFPresignedTTLMin    int     `json:"cf_presigned_ttl_min"`
	QuotaFreeGB          int     `json:"quota_free_gb"`
	QuotaBronzeGB        int     `json:"quota_bronze_gb"`
	QuotaSilverGB        int     `json:"quota_silver_gb"`
	QuotaGoldGB          int     `json:"quota_gold_gb"`
	QuotaPlatinumGB      int     `json:"quota_platinum_gb"`
	Boost1GB             int     `json:"boost_1_gb"`
	Boost1Price          float64 `json:"boost_1_price"`
	Boost2GB             int     `json:"boost_2_gb"`
	Boost2Price          float64 `json:"boost_2_price"`
	Boost3GB             int     `json:"boost_3_gb"`
	Boost3Price          float64 `json:"boost_3_price"`
	ContactEmail         string  `json:"contact_email"`
	HideMatureDefault    bool    `json:"hide_mature_default"`
	EnableComments       bool    `json:"enable_comments"`
	DefaultMetaDescription string `json:"default_meta_description"`
}

var defaultPreferences = UserPreferences{
	Language:            "en",
	ContentLanguage:     "en",
	ItemsPerPage:        12,
	PopularTagsLimit:    20,
	EmailFromFollowing:  true,
	EmailSupportReplies: true,
	EmailMarketing:      false,
	InAppEnabled:        true,
	HideMature:          false,
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
		Language:            settings.DefaultLanguage,
		ContentLanguage:     settings.DefaultContentLang,
		ItemsPerPage:        settings.ItemsPerPage,
		PopularTagsLimit:    settings.PopularTagsLimit,
		EmailFromFollowing:  true,
		EmailSupportReplies: true,
		EmailMarketing:      false,
		InAppEnabled:        true,
		HideMature:          false,
	}, nil
}

// ----- Admin Settings -----

//encore:api auth method=GET path=/admin/settings
func GetAdminSettings(ctx context.Context) (*AppSettings, error) {
	ad := auth.Data().(*AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var raw []byte
	err := db.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = 'defaults'`).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return defaultAppSettings(), nil
	}

	var settings AppSettings
	json.Unmarshal(raw, &settings)

	// Migrate old boost keys to indexed format
	if settings.Boost1GB == 0 && settings.Boost1Price == 0 {
		var old struct {
			Boost5Price  float64 `json:"boost_5gb_price"`
			Boost10Price float64 `json:"boost_10gb_price"`
			Boost20Price float64 `json:"boost_20gb_price"`
		}
		if json.Unmarshal(raw, &old) == nil && old.Boost5Price > 0 {
			settings.Boost1GB = 5
			settings.Boost1Price = old.Boost5Price
			settings.Boost2GB = 10
			settings.Boost2Price = old.Boost10Price
			settings.Boost3GB = 20
			settings.Boost3Price = old.Boost20Price
		}
	}

	return &settings, nil
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
		MaxUploadSizeMB:    50,
		ImageServingMode:   "direct",
		RateLimit:           60,
		S3PresignedTTLMin:   15,
		CFPresignedTTLMin:   15,
		QuotaFreeGB:         1,
		QuotaBronzeGB:       10,
		QuotaSilverGB:       50,
		QuotaGoldGB:         200,
		QuotaPlatinumGB:     1000,
		Boost1GB:             5,
		Boost1Price:          5,
		Boost2GB:             10,
		Boost2Price:          8,
		Boost3GB:             20,
		Boost3Price:          12,
		ContactEmail:         "",
		HideMatureDefault:    false,
		EnableComments:       true,
		DefaultMetaDescription: "",
	}
}
