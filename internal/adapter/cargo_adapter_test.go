package adapter_test

import (
	"context"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/port"
)

func TestCargoAdapter_SourceType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return cargo",
			want: "cargo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter.NewCargoAdapter()
			if got := a.SourceType(); got != tt.want {
				t.Errorf("SourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCargoAdapter_Download_InvalidSource(t *testing.T) {
	adapter := adapter.NewCargoAdapter()
	ctx := context.Background()

	tests := []struct {
		source  *port.Source
		name    string
		wantErr bool
	}{
		{
			source: &port.Source{
				Type: "",
				URL:  "example-crate",
			},
			name:    "empty_type",
			wantErr: true,
		},
		{
			name: "empty_url",
			source: &port.Source{
				Type: "cargo",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "wrong_type",
			source: &port.Source{
				Type: "npm",
				URL:  "example-crate",
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

func TestCargoAdapter_GetLatestVersion_InvalidSource(t *testing.T) {
	adapter := adapter.NewCargoAdapter()
	ctx := context.Background()

	tests := []struct {
		source  *port.Source
		name    string
		wantErr bool
	}{
		{
			source: &port.Source{
				Type: "",
				URL:  "example-crate",
			},
			name:    "empty_type",
			wantErr: true,
		},
		{
			name: "wrong_type",
			source: &port.Source{
				Type: "npm",
				URL:  "example-crate",
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
