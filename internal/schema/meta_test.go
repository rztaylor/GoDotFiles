package schema

import (
	"testing"
)

func TestTypeMeta_ParseKind(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		wantType    string
		wantVersion string
		wantErr     bool
	}{
		{"Valid App", "App/v1", "App", "v1", false},
		{"Valid Profile", "Profile/v2", "Profile", "v2", false},
		{"Missing Version", "App", "", "", true},
		{"Empty String", "", "", "", true},
		{"Too Many Parts", "App/v1/extra", "", "", true},
		{"Empty Type", "/v1", "", "", true},
		{"No 'v' Prefix", "App/1", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := TypeMeta{Kind: tt.kind}
			gotType, gotVersion, err := tm.ParseKind()
			if (err != nil) != tt.wantErr {
				t.Errorf("TypeMeta.ParseKind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotType != tt.wantType {
				t.Errorf("TypeMeta.ParseKind() gotType = %v, want %v", gotType, tt.wantType)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("TypeMeta.ParseKind() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestTypeMeta_ValidateKind(t *testing.T) {
	tests := []struct {
		name         string
		kind         string
		expectedType string
		wantErr      bool
	}{
		{"Match App v1", "App/v1", "App", false},
		{"Mismatch Type", "Profile/v1", "App", true},
		{"Unsupported Version", "App/v2", "App", true},
		{"Invalid Format", "App", "App", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := TypeMeta{Kind: tt.kind}
			if err := tm.ValidateKind(tt.expectedType); (err != nil) != tt.wantErr {
				t.Errorf("TypeMeta.ValidateKind() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
