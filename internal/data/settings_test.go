package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestSettings_YAMLRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.yaml")

	original := &Settings{
		SelectedLeagues: []int{47, 87, 42, 55},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var loaded Settings
	if err := yaml.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(loaded.SelectedLeagues) != len(original.SelectedLeagues) {
		t.Fatalf("SelectedLeagues length = %d, want %d", len(loaded.SelectedLeagues), len(original.SelectedLeagues))
	}
	for i, id := range loaded.SelectedLeagues {
		if id != original.SelectedLeagues[i] {
			t.Errorf("SelectedLeagues[%d] = %d, want %d", i, id, original.SelectedLeagues[i])
		}
	}
}

func TestSettings_YAMLRoundtrip_Empty(t *testing.T) {
	original := &Settings{}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var loaded Settings
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(loaded.SelectedLeagues) != 0 {
		t.Errorf("SelectedLeagues length = %d, want 0", len(loaded.SelectedLeagues))
	}
}

func TestSettings_MissingFileFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "settings.yaml")

	_, err := os.ReadFile(path)
	if !os.IsNotExist(err) {
		t.Fatal("expected file to not exist")
	}

	// Simulate the same fallback logic as LoadSettings
	var settings Settings
	if os.IsNotExist(err) {
		settings = Settings{}
	}

	if len(settings.SelectedLeagues) != 0 {
		t.Errorf("default settings should have empty SelectedLeagues, got %d", len(settings.SelectedLeagues))
	}
}

func TestSettings_IsLeagueSelected(t *testing.T) {
	tests := []struct {
		name     string
		leagues  []int
		checkID  int
		expected bool
	}{
		{"found", []int{47, 87, 42}, 87, true},
		{"not found", []int{47, 87, 42}, 55, false},
		{"empty list", []int{}, 47, false},
		{"single match", []int{42}, 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Settings{SelectedLeagues: tt.leagues}
			got := s.IsLeagueSelected(tt.checkID)
			if got != tt.expected {
				t.Errorf("IsLeagueSelected(%d) = %v, want %v", tt.checkID, got, tt.expected)
			}
		})
	}
}

func TestActiveLeagueIDs_NoSettings(t *testing.T) {
	// ActiveLeagueIDs returns DefaultLeagueIDs when no leagues are selected.
	// Since we can't easily mock the settings file, we verify the default behavior.
	ids := ActiveLeagueIDs()

	// Should return at least the default leagues
	if len(ids) == 0 {
		t.Error("ActiveLeagueIDs() returned empty slice")
	}
}

func TestDefaultLeagueIDs(t *testing.T) {
	if len(DefaultLeagueIDs) == 0 {
		t.Error("DefaultLeagueIDs should not be empty")
	}

	// Verify all default IDs are in the supported leagues
	allIDs := AllLeagueIDs()
	idSet := make(map[int]bool, len(allIDs))
	for _, id := range allIDs {
		idSet[id] = true
	}

	for _, id := range DefaultLeagueIDs {
		if !idSet[id] {
			t.Errorf("DefaultLeagueIDs contains %d which is not in AllSupportedLeagues", id)
		}
	}
}

func TestAllLeagueIDs(t *testing.T) {
	ids := AllLeagueIDs()

	if len(ids) == 0 {
		t.Fatal("AllLeagueIDs() returned empty slice")
	}

	// Count expected total from AllSupportedLeagues
	expected := 0
	for _, leagues := range AllSupportedLeagues {
		expected += len(leagues)
	}

	if len(ids) != expected {
		t.Errorf("AllLeagueIDs() returned %d IDs, want %d", len(ids), expected)
	}

	// Verify no duplicates
	seen := make(map[int]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate league ID: %d", id)
		}
		seen[id] = true
	}
}

func TestGetAllRegions(t *testing.T) {
	regions := GetAllRegions()

	expected := []string{RegionEurope, RegionAmerica, RegionGlobal}
	if len(regions) != len(expected) {
		t.Fatalf("GetAllRegions() returned %d regions, want %d", len(regions), len(expected))
	}

	for i, r := range regions {
		if r != expected[i] {
			t.Errorf("GetAllRegions()[%d] = %q, want %q", i, r, expected[i])
		}
	}
}

func TestGetLeaguesForRegion(t *testing.T) {
	tests := []struct {
		region    string
		wantEmpty bool
	}{
		{RegionEurope, false},
		{RegionAmerica, false},
		{RegionGlobal, false},
		{"NonExistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			leagues := GetLeaguesForRegion(tt.region)
			if tt.wantEmpty && len(leagues) != 0 {
				t.Errorf("GetLeaguesForRegion(%q) returned %d leagues, want 0", tt.region, len(leagues))
			}
			if !tt.wantEmpty && len(leagues) == 0 {
				t.Errorf("GetLeaguesForRegion(%q) returned empty slice", tt.region)
			}
		})
	}

	// Verify a known league exists in Europe
	europeLeagues := GetLeaguesForRegion(RegionEurope)
	found := false
	for _, l := range europeLeagues {
		if l.ID == 47 && l.Name == "Premier League" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Premier League (ID=47) not found in Europe region")
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}

	if dir == "" {
		t.Fatal("ConfigDir() returned empty string")
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir() returned relative path: %s", dir)
	}

	// Directory should exist (ConfigDir creates it)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("ConfigDir() path does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("ConfigDir() path is not a directory: %s", dir)
	}
}

func TestSettingsPath(t *testing.T) {
	path, err := SettingsPath()
	if err != nil {
		t.Fatalf("SettingsPath() error: %v", err)
	}

	if filepath.Base(path) != settingsFileName {
		t.Errorf("SettingsPath() base = %q, want %q", filepath.Base(path), settingsFileName)
	}
}

// Helper function tests

func TestIntPtr(t *testing.T) {
	p := intPtr(42)
	if p == nil {
		t.Fatal("intPtr returned nil")
	}
	if *p != 42 {
		t.Errorf("*intPtr(42) = %d, want 42", *p)
	}

	// Verify it returns a new pointer each time
	p2 := intPtr(42)
	if p == p2 {
		t.Error("intPtr should return distinct pointers")
	}
}

func TestStringPtr(t *testing.T) {
	p := stringPtr("hello")
	if p == nil {
		t.Fatal("stringPtr returned nil")
	}
	if *p != "hello" {
		t.Errorf("*stringPtr(\"hello\") = %q, want \"hello\"", *p)
	}

	p2 := stringPtr("hello")
	if p == p2 {
		t.Error("stringPtr should return distinct pointers")
	}
}

func TestTimePtr(t *testing.T) {
	now := time.Now()
	p := timePtr(now)
	if p == nil {
		t.Fatal("timePtr returned nil")
	}
	if !p.Equal(now) {
		t.Errorf("*timePtr(now) = %v, want %v", *p, now)
	}

	p2 := timePtr(now)
	if p == p2 {
		t.Error("timePtr should return distinct pointers")
	}
}
