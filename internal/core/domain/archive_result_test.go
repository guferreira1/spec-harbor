package domain

import "testing"

func TestNewArchiveResultStoresArchiveDetails(t *testing.T) {
	movedDirectory := ArchiveMovedDirectory{
		SourcePath:  "openspec/changes/change",
		ArchivePath: "openspec/archive/2026-06-06/change",
	}

	result := NewArchiveResult(
		"change",
		"openspec/changes/change",
		"openspec/archive/2026-06-06/change",
		"2026-06-06",
		movedDirectory,
	)

	if result.ChangeID != "change" {
		t.Fatalf("ChangeID = %q, want change", result.ChangeID)
	}
	if result.SourcePath != "openspec/changes/change" {
		t.Fatalf("SourcePath = %q, want openspec/changes/change", result.SourcePath)
	}
	if result.ArchivePath != "openspec/archive/2026-06-06/change" {
		t.Fatalf("ArchivePath = %q, want openspec/archive/2026-06-06/change", result.ArchivePath)
	}
	if result.ArchiveDate != "2026-06-06" {
		t.Fatalf("ArchiveDate = %q, want 2026-06-06", result.ArchiveDate)
	}
	if result.MovedDirectory != movedDirectory {
		t.Fatalf("MovedDirectory = %#v, want %#v", result.MovedDirectory, movedDirectory)
	}
	if !result.Moved() {
		t.Fatalf("Moved() = false, want true")
	}
}

func TestArchiveResultMovedRequiresMovedDirectoryPaths(t *testing.T) {
	if (ArchiveResult{}).Moved() {
		t.Fatalf("zero ArchiveResult Moved() = true, want false")
	}

	result := ArchiveResult{MovedDirectory: ArchiveMovedDirectory{SourcePath: "source"}}
	if result.Moved() {
		t.Fatalf("partial ArchiveResult Moved() = true, want false")
	}
}
