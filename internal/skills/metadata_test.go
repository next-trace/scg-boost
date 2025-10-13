package skills

import (
	"testing"
	"testing/fstest"
)

func TestLoadMetadata(t *testing.T) {
	tests := []struct {
		name      string
		fs        fstest.MapFS
		skillPath string
		wantName  string
		wantErr   bool
	}{
		{
			name: "valid metadata",
			fs: fstest.MapFS{
				"gateway-service/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "gateway-service",
						"name": "gateway-service",
						"type": "concrete",
						"version": "1.0.0",
						"description": "Gateway service",
						"tags": ["service", "http"],
						"repo_types": ["go-service"],
						"author": "SupplyChainGuard"
					}`),
				},
			},
			skillPath: "gateway-service",
			wantName:  "gateway-service",
			wantErr:   false,
		},
		{
			name: "missing name",
			fs: fstest.MapFS{
				"bad/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "bad",
						"type": "generic",
						"version": "1.0.0",
						"description": "Bad skill",
						"repo_types": ["generic"]
					}`),
				},
			},
			skillPath: "bad",
			wantErr:   true,
		},
		{
			name: "missing version",
			fs: fstest.MapFS{
				"bad/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "bad",
						"name": "bad",
						"type": "generic",
						"description": "Bad skill",
						"repo_types": ["generic"]
					}`),
				},
			},
			skillPath: "bad",
			wantErr:   true,
		},
		{
			name:      "file not found",
			fs:        fstest.MapFS{},
			skillPath: "missing",
			wantErr:   true,
		},
		{
			name: "invalid json",
			fs: fstest.MapFS{
				"bad/skill.json": &fstest.MapFile{
					Data: []byte(`not json`),
				},
			},
			skillPath: "bad",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := LoadMetadata(tt.fs, tt.skillPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && meta.Name != tt.wantName {
				t.Errorf("LoadMetadata() name = %v, want %v", meta.Name, tt.wantName)
			}
		})
	}
}
