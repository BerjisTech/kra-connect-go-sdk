package kra

import (
	"testing"
	"time"
)

func TestValidateAndNormalizePIN(t *testing.T) {
	tests := []struct {
		name    string
		pin     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid PIN uppercase",
			pin:     "P051234567A",
			want:    "P051234567A",
			wantErr: false,
		},
		{
			name:    "valid PIN lowercase",
			pin:     "p051234567a",
			want:    "P051234567A",
			wantErr: false,
		},
		{
			name:    "valid PIN with spaces",
			pin:     "  P051234567A  ",
			want:    "P051234567A",
			wantErr: false,
		},
		{
			name:    "empty PIN",
			pin:     "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - missing P",
			pin:     "051234567A",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - too short",
			pin:     "P05123456A",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - too long",
			pin:     "P0512345678A",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - missing letter",
			pin:     "P0512345678",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - letter in middle",
			pin:     "P051234A567",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndNormalizePIN(tt.pin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndNormalizePIN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateAndNormalizePIN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndNormalizeTCC(t *testing.T) {
	tests := []struct {
		name    string
		tcc     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid TCC uppercase",
			tcc:     "TCC123456",
			want:    "TCC123456",
			wantErr: false,
		},
		{
			name:    "valid TCC lowercase",
			tcc:     "tcc123456",
			want:    "TCC123456",
			wantErr: false,
		},
		{
			name:    "valid TCC with spaces",
			tcc:     "  TCC123456  ",
			want:    "TCC123456",
			wantErr: false,
		},
		{
			name:    "empty TCC",
			tcc:     "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - missing TCC prefix",
			tcc:     "123456",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - letters after TCC",
			tcc:     "TCCABC123",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndNormalizeTCC(tt.tcc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndNormalizeTCC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateAndNormalizeTCC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEslipNumber(t *testing.T) {
	tests := []struct {
		name    string
		eslip   string
		wantErr bool
	}{
		{
			name:    "valid e-slip",
			eslip:   "1234567890",
			wantErr: false,
		},
		{
			name:    "valid e-slip with spaces",
			eslip:   "  1234567890  ",
			wantErr: false,
		},
		{
			name:    "empty e-slip",
			eslip:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - letters",
			eslip:   "12345ABC890",
			wantErr: true,
		},
		{
			name:    "invalid format - special characters",
			eslip:   "1234-5678-90",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEslipNumber(tt.eslip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEslipNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePeriod(t *testing.T) {
	tests := []struct {
		name    string
		period  string
		wantErr bool
	}{
		{
			name:    "valid period",
			period:  "202401",
			wantErr: false,
		},
		{
			name:    "valid period - December",
			period:  "202312",
			wantErr: false,
		},
		{
			name:    "empty period",
			period:  "",
			wantErr: true,
		},
		{
			name:    "invalid format - too short",
			period:  "20241",
			wantErr: true,
		},
		{
			name:    "invalid format - too long",
			period:  "2024011",
			wantErr: true,
		},
		{
			name:    "invalid month - 00",
			period:  "202400",
			wantErr: true,
		},
		{
			name:    "invalid month - 13",
			period:  "202413",
			wantErr: true,
		},
		{
			name:    "invalid year - too low",
			period:  "189912",
			wantErr: true,
		},
		{
			name:    "invalid year - too high",
			period:  "210112",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePeriod(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePeriod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  "1234567890123456",
			wantErr: false,
		},
		{
			name:    "valid long API key",
			apiKey:  "very-long-api-key-with-lots-of-characters",
			wantErr: false,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "too short API key",
			apiKey:  "short",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "valid timeout",
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "valid max timeout",
			timeout: 10 * time.Minute,
			wantErr: false,
		},
		{
			name:    "zero timeout",
			timeout: 0,
			wantErr: true,
		},
		{
			name:    "negative timeout",
			timeout: -5 * time.Second,
			wantErr: true,
		},
		{
			name:    "too large timeout",
			timeout: 15 * time.Minute,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeout(tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRetryConfig(t *testing.T) {
	tests := []struct {
		name         string
		maxRetries   int
		initialDelay time.Duration
		maxDelay     time.Duration
		wantErr      bool
	}{
		{
			name:         "valid config",
			maxRetries:   3,
			initialDelay: 1 * time.Second,
			maxDelay:     32 * time.Second,
			wantErr:      false,
		},
		{
			name:         "negative retries",
			maxRetries:   -1,
			initialDelay: 1 * time.Second,
			maxDelay:     32 * time.Second,
			wantErr:      true,
		},
		{
			name:         "too many retries",
			maxRetries:   15,
			initialDelay: 1 * time.Second,
			maxDelay:     32 * time.Second,
			wantErr:      true,
		},
		{
			name:         "zero initial delay",
			maxRetries:   3,
			initialDelay: 0,
			maxDelay:     32 * time.Second,
			wantErr:      true,
		},
		{
			name:         "max delay less than initial",
			maxRetries:   3,
			initialDelay: 10 * time.Second,
			maxDelay:     5 * time.Second,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRetryConfig(tt.maxRetries, tt.initialDelay, tt.maxDelay)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRetryConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRateLimitConfig(t *testing.T) {
	tests := []struct {
		name        string
		maxRequests int
		window      time.Duration
		wantErr     bool
	}{
		{"valid", 10, time.Second, false},
		{"zero requests", 0, time.Second, true},
		{"negative requests", -5, time.Second, true},
		{"zero window", 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRateLimitConfig(tt.maxRequests, tt.window)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRateLimitConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateObligationID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "OBL123456", false},
		{"empty", "", true},
		{"invalid characters", "OBL 123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObligationID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObligationID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
