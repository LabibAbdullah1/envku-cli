package installer

import (
	"testing"
)

func TestDistroCodenameMapping(t *testing.T) {
	tests := []struct {
		name     string
		distro   DistroInfo
		expected string
		isDebian bool
	}{
		{
			name: "Linux Mint 22 (Zena / Wilma)",
			distro: DistroInfo{
				ID:       "linuxmint",
				IDLike:   "ubuntu debian",
				Name:     "Linux Mint",
				Codename: "zena",
			},
			expected: "noble",
			isDebian: false,
		},
		{
			name: "Linux Mint 21 (Virginia / Victoria)",
			distro: DistroInfo{
				ID:       "linuxmint",
				IDLike:   "ubuntu debian",
				Name:     "Linux Mint",
				Codename: "virginia",
			},
			expected: "jammy",
			isDebian: false,
		},
		{
			name: "Ubuntu 24.04 (Noble)",
			distro: DistroInfo{
				ID:       "ubuntu",
				Codename: "noble",
			},
			expected: "noble",
			isDebian: false,
		},
		{
			name: "Pop!_OS 22.04",
			distro: DistroInfo{
				ID:        "pop",
				IDLike:    "ubuntu debian",
				VersionID: "22.04",
				Codename:  "jammy",
			},
			expected: "jammy",
			isDebian: false,
		},
		{
			name: "Debian 12 (Bookworm)",
			distro: DistroInfo{
				ID:       "debian",
				Codename: "bookworm",
			},
			expected: "bookworm",
			isDebian: true,
		},
		{
			name: "LMDE 6 (Faye - Debian Based)",
			distro: DistroInfo{
				ID:       "linuxmint",
				IDLike:   "debian",
				Name:     "LMDE 6 Faye",
				Codename: "faye",
			},
			expected: "bookworm",
			isDebian: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isDebian {
				if !tt.distro.IsDebian() {
					t.Errorf("expected IsDebian() = true")
				}
				got := tt.distro.GetEffectiveDebianCodename()
				if got != tt.expected {
					t.Errorf("GetEffectiveDebianCodename() = %s; want %s", got, tt.expected)
				}
			} else {
				if tt.distro.IsDebian() {
					t.Errorf("expected IsDebian() = false")
				}
				got := tt.distro.GetEffectiveUbuntuCodename()
				if got != tt.expected {
					t.Errorf("GetEffectiveUbuntuCodename() = %s; want %s", got, tt.expected)
				}
			}
		})
	}
}

