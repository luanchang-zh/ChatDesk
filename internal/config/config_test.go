package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		name      string
		env       map[string]string
		wantError string
		wantAddr  string
		wantUser  string
	}{
		{name: "identity disabled by default", wantAddr: ":8080"},
		{name: "development defaults to loopback", env: map[string]string{"DEV_AUTH_ENABLED": "true", "DEV_USER_ID": "A525E8F9-1F6A-47B4-BE50-C7CE5AFDE696"}, wantAddr: "127.0.0.1:8080", wantUser: "a525e8f9-1f6a-47b4-be50-c7ce5afde696"},
		{name: "IPv6 loopback", env: map[string]string{"DEV_AUTH_ENABLED": "true", "DEV_USER_ID": "a525e8f9-1f6a-47b4-be50-c7ce5afde696", "HTTP_ADDR": "[::1]:8080"}, wantAddr: "[::1]:8080", wantUser: "a525e8f9-1f6a-47b4-be50-c7ce5afde696"},
		{name: "missing database", env: map[string]string{"POSTGRES_DSN": ""}, wantError: "POSTGRES_DSN"},
		{name: "bad flag", env: map[string]string{"DEV_AUTH_ENABLED": "maybe"}, wantError: "DEV_AUTH_ENABLED"},
		{name: "missing identity", env: map[string]string{"DEV_AUTH_ENABLED": "true"}, wantError: "DEV_USER_ID"},
		{name: "zero identity", env: map[string]string{"DEV_AUTH_ENABLED": "true", "DEV_USER_ID": "00000000-0000-0000-0000-000000000000"}, wantError: "DEV_USER_ID"},
		{name: "all interfaces", env: map[string]string{"DEV_AUTH_ENABLED": "true", "HTTP_ADDR": ":8080"}, wantError: "回环"},
		{name: "external interface", env: map[string]string{"DEV_AUTH_ENABLED": "true", "HTTP_ADDR": "192.0.2.1:8080"}, wantError: "回环"},
		{name: "hostnames cannot change resolution", env: map[string]string{"DEV_AUTH_ENABLED": "true", "HTTP_ADDR": "localhost:8080"}, wantError: "回环"},
		{name: "bad port", env: map[string]string{"HTTP_ADDR": "127.0.0.1:65536"}, wantError: "端口"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range map[string]string{"POSTGRES_DSN": "postgres://user:canary-secret@localhost/db", "HTTP_ADDR": "", "LOG_LEVEL": "", "DEV_AUTH_ENABLED": "", "DEV_USER_ID": ""} {
				t.Setenv(key, value)
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			cfg, err := Load()
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) || strings.Contains(err.Error(), "canary-secret") {
					t.Fatalf("expected safe error containing %q, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil || cfg.HTTPAddr != tc.wantAddr || cfg.DevUserID != tc.wantUser {
				t.Fatalf("unexpected config: address=%q user=%q error=%v", cfg.HTTPAddr, cfg.DevUserID, err)
			}
		})
	}
}
