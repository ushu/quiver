package quiver_test

import (
	"path/filepath"
	"testing"

	"github.com/ushu/quiver"
)

func TestLoadLibrary(t *testing.T) {
	t.Parallel()
	libPath := fixturePath("Quiver.qvlibrary")

	// Ensure we can load the library
	lib, err := quiver.ReadLibrary(libPath, false)
	if err != nil {
		t.Error(err)
	}

	// It should have one notebook
	if len(lib.Notebooks) != 1 {
		t.Errorf("len(lib.Notebooks) = %v; want %v", len(lib.Notebooks), 1)
	}
}

func TestLoadNotebook(t *testing.T) {
	t.Parallel()
	nbPath := fixturePath("Quiver.qvlibrary/Quiver Test.qvnotebook")

	// Ensure we can load the notebook
	nb, err := quiver.ReadNotebook(nbPath, false)
	if err != nil {
		t.Error(err)
	}

	// validate metadata
	const name = "Quiver Test"
	if nb.Name != name {
		t.Errorf("nb.Name = %q, want %q", nb.Name, name)
	}
	const UUID = "FIXTURE"
	if nb.UUID != UUID {
		t.Errorf("nb.UUID = %q; want %q", nb.UUID, UUID)
	}

	// It should have 3 notes
	if len(nb.Notes) != 3 {
		t.Errorf("len(nb.Notes) = %v; want %v", len(nb.Notes), 3)
	}
}

func TestLoadNoteWithTags(t *testing.T) {
	t.Parallel()
	notePath := fixturePath("Quiver.qvlibrary/Quiver Test.qvnotebook/73385592-0CAB-41E5-9045-AEC528C2915A.qvnote")

	// Ensure we can load the notebook
	note, err := quiver.ReadNote(notePath, false)
	if err != nil {
		t.Error(err)
	}

	// validate metadata
	const title = "Tags"
	if note.Title != title {
		t.Errorf("note.Title = %q, want %q", note.Title, title)
	}
	var tags = []string{"retest", "tags", "test"}
	if !stringSliceEqual(note.Tags, tags) {
		t.Errorf("note.Tags = %q, want %q", note.Tags, tags)
	}
	const UUID = "73385592-0CAB-41E5-9045-AEC528C2915A"
	if note.UUID != UUID {
		t.Errorf("note.UUID = %q; want %q", note.UUID, UUID)
	}

	// It should have 1 cell
	if len(note.Cells) != 1 {
		t.Errorf("len(note.Cells) = %v; want %v", len(note.Cells), 1)
	}
}

func TestLoadNoteSeveralResources(t *testing.T) {
	t.Parallel()
	notePath := fixturePath("Quiver.qvlibrary/Quiver Test.qvnotebook/B59AC519-2A2C-4EC8-B701-E69F54F40A85.qvnote")

	// Ensure we can load the note
	note, err := quiver.ReadNote(notePath, true)
	if err != nil {
		t.Error(err)
	}

	// validate metadata
	const title = "Images, Files and Links"
	if note.Title != title {
		t.Errorf("note.Title = %q, want %q", note.Title, title)
	}
	var tags = []string{}
	if !stringSliceEqual(note.Tags, tags) {
		t.Errorf("note.Tags = %q, want %q", note.Tags, tags)
	}
	const UUID = "B59AC519-2A2C-4EC8-B701-E69F54F40A85"
	if note.UUID != UUID {
		t.Errorf("note.UUID = %q; want %q", note.UUID, UUID)
	}

	// resources
	if len(note.Resources) != 2 {
		t.Errorf("len(Note.Resources) = %v; want %v", len(note.Resources), 2)
	} else {
		resNames := []string{
			"1C3392AA-54E7-4EA3-A129-1C20F208B029.jpg",
			"F6E1CA4A-FA0B-4E45-9861-3E3FEB0DAF99.png",
		}
		for i := range note.Resources {
			if note.Resources[i].Name != resNames[i] {
				t.Errorf("len(note.Resource[%v].Name) = %q; want %q", i, note.Resources[i].Name, resNames[i])
			}
		}
	}
}

func TestLoadNoteSeveralCells(t *testing.T) {
	t.Parallel()
	notePath := fixturePath("Quiver.qvlibrary/Quiver Test.qvnotebook/D2A1CC36-CC97-4701-A895-EFC98EF47026.qvnote")

	// Ensure we can load the note
	note, err := quiver.ReadNote(notePath, false)
	if err != nil {
		t.Error(err)
	}

	// validate metadata
	const title = "Text cells"
	if note.Title != title {
		t.Errorf("note.Title = %q, want %q", note.Title, title)
	}
	var tags = []string{"tutorial"}
	if !stringSliceEqual(note.Tags, tags) {
		t.Errorf("note.Tags = %q, want %q", note.Tags, tags)
	}
	const UUID = "D2A1CC36-CC97-4701-A895-EFC98EF47026"
	if note.UUID != UUID {
		t.Errorf("note.UUID = %q; want %q", note.UUID, UUID)
	}

	// It should have 3 cell
	if len(note.Cells) != 3 {
		t.Errorf("len(note.Cells) = %v; want %v", len(note.Cells), 3)
	}
}

// Helpers

func fixturePath(p string) string {
	return filepath.Join("testdata", p)
}

func stringSliceEqual(l []string, r []string) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}
