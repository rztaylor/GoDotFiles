package apps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveName(t *testing.T) {
	tests := []struct {
		name        string
		pkg         *Package
		manager     string
		wantName    string
		wantDefined bool
	}{
		{
			name: "All defined, request brew",
			pkg: &Package{
				Brew: "brew-pkg",
				Apt:  &AptPackage{Name: "apt-pkg"},
				Dnf:  "dnf-pkg",
			},
			manager:     "brew",
			wantName:    "brew-pkg",
			wantDefined: true,
		},
		{
			name: "All defined, request apt",
			pkg: &Package{
				Brew: "brew-pkg",
				Apt:  &AptPackage{Name: "apt-pkg"},
			},
			manager:     "apt",
			wantName:    "apt-pkg",
			wantDefined: true,
		},
		{
			name: "Only brew defined, request apt (should be empty)",
			pkg: &Package{
				Brew: "brew-pkg",
			},
			manager:     "apt",
			wantName:    "",
			wantDefined: false,
		},
		{
			name: "Only apt defined, request brew (should be empty)",
			pkg: &Package{
				Apt: &AptPackage{Name: "apt-pkg"},
			},
			manager:     "brew",
			wantName:    "",
			wantDefined: false,
		},
		{
			name:        "No package defined",
			pkg:         &Package{},
			manager:     "apt",
			wantName:    "",
			wantDefined: false,
		},
		{
			name: "Request unknown manager",
			pkg: &Package{
				Brew: "foo",
			},
			manager:     "unknown",
			wantName:    "",
			wantDefined: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotDefined := tt.pkg.ResolveName(tt.manager)
			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantDefined, gotDefined)
		})
	}
}
