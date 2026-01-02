package data

import "testing"

func TestEmbeddedThemeRepository_LoadsThemes(t *testing.T) {
	repo, err := NewEmbeddedThemeRepository()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	theme, err := repo.GetThemeByName("birds")
	if err != nil {
		t.Fatalf("expected birds theme, got error: %v", err)
	}
	if len(theme.Items) == 0 {
		t.Fatalf("expected birds theme to have items")
	}
}

func TestEmbeddedThemeRepository_GetAllThemeNames(t *testing.T) {
	repo, err := NewEmbeddedThemeRepository()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := repo.GetAllThemeNames()
	if len(names) < 3 {
		t.Fatalf("expected at least 3 themes, got %d", len(names))
	}

	foundBirds := false
	for _, name := range names {
		if name == "birds" {
			foundBirds = true
			break
		}
	}
	if !foundBirds {
		t.Fatalf("expected birds theme to be listed")
	}
}

func TestFilterItems_Exclude(t *testing.T) {
	repo, err := NewEmbeddedThemeRepository()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	theme, err := repo.GetThemeByName("birds")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filtered := FilterItems(theme.Items, []string{"blue-heron", "albatross"})
	for _, item := range filtered {
		if normalizeName(item.Name) == "blue-heron" || normalizeName(item.Name) == "albatross" {
			t.Fatalf("expected item %q to be filtered out", item.Name)
		}
	}
}
