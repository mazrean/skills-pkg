package adapter_test

import (
	"context"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/port"
)

func TestPipAdapter_SourceType(t *testing.T) {
	adapter := adapter.NewPipAdapter()
	if got := adapter.SourceType(); got != "pip" {
		t.Errorf("SourceType() = %v, want %v", got, "pip")
	}
}

func TestPipAdapter_Download_InvalidSource(t *testing.T) {
	adapter := adapter.NewPipAdapter()
	ctx := context.Background()

	tests := []struct {
		source  *port.Source
		name    string
		wantErr bool
	}{
		{
			source: &port.Source{
				Type: "",
				URL:  "example-package",
			},
			name:    "empty_type",
			wantErr: true,
		},
		{
			name: "empty_url",
			source: &port.Source{
				Type: "pip",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "wrong_type",
			source: &port.Source{
				Type: "npm",
				URL:  "example-package",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adapter.Download(ctx, tt.source, "latest")
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipAdapter_GetLatestVersion_InvalidSource(t *testing.T) {
	adapter := adapter.NewPipAdapter()
	ctx := context.Background()

	tests := []struct {
		source  *port.Source
		name    string
		wantErr bool
	}{
		{
			source: &port.Source{
				Type: "",
				URL:  "example-package",
			},
			name:    "empty_type",
			wantErr: true,
		},
		{
			name: "wrong_type",
			source: &port.Source{
				Type: "npm",
				URL:  "example-package",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adapter.GetLatestVersion(ctx, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
