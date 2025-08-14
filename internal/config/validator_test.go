package config

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid minimal config",
			config: &Config{
				Version: "1.0",
				Settings: Settings{
					BaseDir:             "/tmp/agents",
					ConflictStrategy:    "backup",
					LogLevel:            "info",
					ConcurrentDownloads: 3,
				},
				Sources: []Source{
					{
						Name:       "test",
						Type:       "github",
						Repository: "user/repo",
						Paths: PathConfig{
							Source: "src",
							Target: "/tmp/test",
						},
					},
				},
				Metadata: Metadata{
					TrackingFile: "/tmp/tracking.json",
					LogFile:      "/tmp/agent-manager.log",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid source type",
			config: &Config{
				Settings: Settings{
					BaseDir: "/tmp/agents",
				},
				Sources: []Source{
					{
						Name:       "test",
						Type:       "invalid",
						Repository: "user/repo",
						Paths: PathConfig{
							Source: "src",
							Target: "/tmp/test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing repository for github source",
			config: &Config{
				Settings: Settings{
					BaseDir: "/tmp/agents",
				},
				Sources: []Source{
					{
						Name: "test",
						Type: "github",
						Paths: PathConfig{
							Source: "src",
							Target: "/tmp/test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty source name",
			config: &Config{
				Settings: Settings{
					BaseDir: "/tmp/agents",
				},
				Sources: []Source{
					{
						Name:       "",
						Type:       "github",
						Repository: "user/repo",
						Paths: PathConfig{
							Source: "src",
							Target: "/tmp/test",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSource(t *testing.T) {
	tests := []struct {
		name    string
		source  Source
		wantErr bool
	}{
		{
			name: "valid github source",
			source: Source{
				Name:       "test",
				Type:       "github",
				Repository: "user/repo",
				Paths: PathConfig{
					Source: "src",
					Target: "/tmp/test",
				},
			},
			wantErr: false,
		},
		{
			name: "valid git source",
			source: Source{
				Name: "test",
				Type: "git",
				URL:  "https://github.com/user/repo.git",
				Paths: PathConfig{
					Source: "src",
					Target: "/tmp/test",
				},
			},
			wantErr: false,
		},
		{
			name: "valid local source",
			source: Source{
				Name: "test",
				Type: "local",
				Paths: PathConfig{
					Source: "/home/user/src",
					Target: "/tmp/test",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid source type",
			source: Source{
				Name: "test",
				Type: "ftp",
				Paths: PathConfig{
					Source: "src",
					Target: "/tmp/test",
				},
			},
			wantErr: true,
		},
		{
			name: "github source missing repository",
			source: Source{
				Name: "test",
				Type: "github",
				Paths: PathConfig{
					Source: "src",
					Target: "/tmp/test",
				},
			},
			wantErr: true,
		},
		{
			name: "git source missing URL",
			source: Source{
				Name: "test",
				Type: "git",
				Paths: PathConfig{
					Source: "src",
					Target: "/tmp/test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing source path for local source",
			source: Source{
				Name: "test",
				Type: "local",
				Paths: PathConfig{
					Target: "/tmp/test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing target path",
			source: Source{
				Name:       "test",
				Type:       "github",
				Repository: "user/repo",
				Paths: PathConfig{
					Source: "src",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSource(&tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "existing command",
			command: "echo",
			want:    true,
		},
		{
			name:    "non-existing command",
			command: "nonexistentcommand12345",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commandExists(tt.command)
			if got != tt.want {
				t.Errorf("commandExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "regular path",
			path: "/tmp/test",
			want: "/tmp/test",
		},
		{
			name: "relative path",
			path: "test/file",
			want: "test/file",
		},
		// Note: home expansion test would be system-dependent
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.want && tt.name != "home expansion" {
				t.Errorf("expandPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
