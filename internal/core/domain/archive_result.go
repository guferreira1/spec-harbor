package domain

type ArchiveMovedDirectory struct {
	SourcePath  string
	ArchivePath string
}

type ArchiveResult struct {
	ChangeID       string
	SourcePath     string
	ArchivePath    string
	ArchiveDate    string
	MovedDirectory ArchiveMovedDirectory
}

func NewArchiveResult(changeID string, sourcePath string, archivePath string, archiveDate string, movedDirectory ArchiveMovedDirectory) ArchiveResult {
	return ArchiveResult{
		ChangeID:       changeID,
		SourcePath:     sourcePath,
		ArchivePath:    archivePath,
		ArchiveDate:    archiveDate,
		MovedDirectory: movedDirectory,
	}
}

func (result ArchiveResult) Moved() bool {
	return result.MovedDirectory.SourcePath != "" && result.MovedDirectory.ArchivePath != ""
}
