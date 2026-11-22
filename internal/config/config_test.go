package config

import (
	"strings"
	"testing"
	"time"
)

func fakeEnv(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

func TestLoad(t *testing.T) {
	t.Parallel()

	const dbURL = "postgres://paygo:paygo@localhost:5433/paygo?sslmode=disable"

	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr []string // fragments that must all appear in the error; nil means success
	}{
		{
			name: "all values valid",
			env: map[string]string{
				"APP_ENV":          "prod",
				"HTTP_ADDR":        ":9090",
				"DATABASE_URL":     dbURL,
				"SHUTDOWN_TIMEOUT": "30s",
			},
			want: Config{
				Env:             "prod",
				HTTPAddr:        ":9090",
				DatabaseURL:     dbURL,
				ShutdownTimeout: 30 * time.Second,
			},
		},
		{
			name: "defaults apply when only database url is set",
			env:  map[string]string{"DATABASE_URL": dbURL},
			want: Config{
				Env:             "dev",
				HTTPAddr:        ":8080",
				DatabaseURL:     dbURL,
				ShutdownTimeout: 15 * time.Second,
			},
		},
		{
			name: "surrounding whitespace is trimmed",
			env: map[string]string{
				"APP_ENV":      " staging ",
				"DATABASE_URL": "  " + dbURL + " ",
			},
			want: Config{
				Env:             "staging",
				HTTPAddr:        ":8080",
				DatabaseURL:     dbURL,
				ShutdownTimeout: 15 * time.Second,
			},
		},
		{
			name:    "missing database url",
			env:     map[string]string{},
			wantErr: []string{"DATABASE_URL is required"},
		},
		{
			name: "invalid app env",
			env: map[string]string{
				"APP_ENV":      "porduction",
				"DATABASE_URL": dbURL,
			},
			wantErr: []string{"invalid APP_ENV", "porduction"},
		},
		{
			name: "unparseable shutdown timeout",
			env: map[string]string{
				"DATABASE_URL":     dbURL,
				"SHUTDOWN_TIMEOUT": "abc",
			},
			wantErr: []string{"invalid SHUTDOWN_TIMEOUT", "abc"},
		},
		{
			name: "negative shutdown timeout",
			env: map[string]string{
				"DATABASE_URL":     dbURL,
				"SHUTDOWN_TIMEOUT": "-5s",
			},
			wantErr: []string{"SHUTDOWN_TIMEOUT must be positive"},
		},
		{
			name: "zero shutdown timeout",
			env: map[string]string{
				"DATABASE_URL":     dbURL,
				"SHUTDOWN_TIMEOUT": "0s",
			},
			wantErr: []string{"SHUTDOWN_TIMEOUT must be positive"},
		},
		{
			name: "several problems are all reported together",
			env: map[string]string{
				"APP_ENV":          "nope",
				"SHUTDOWN_TIMEOUT": "abc",
			},
			wantErr: []string{
				"invalid APP_ENV",
				"DATABASE_URL is required",
				"invalid SHUTDOWN_TIMEOUT",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Load(fakeEnv(tc.env))

			if len(tc.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load() unexpected error: %v", err)
				}
				if got != tc.want {
					t.Errorf("Load() = %+v, want %+v", got, tc.want)
				}
				return
			}

			if err == nil {
				t.Fatalf("Load() error = nil, want error containing %q", tc.wantErr)
			}
			for _, fragment := range tc.wantErr {
				if !strings.Contains(err.Error(), fragment) {
					t.Errorf("error %q does not contain %q", err.Error(), fragment)
				}
			}
			if got != (Config{}) {
				t.Errorf("Load() returned %+v on error, want zero Config", got)
			}
		})
	}
}
