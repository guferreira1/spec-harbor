package version

import "fmt"

const (
	defaultVersion = "dev"
	defaultCommit  = "unknown"
	defaultDate    = "unknown"
	defaultDirty   = "unknown"
)

var (
	Version = defaultVersion
	Commit  = defaultCommit
	Date    = defaultDate
	Dirty   = defaultDirty
)

type Metadata struct {
	Version string
	Commit  string
	Date    string
	Dirty   string
}

func Current() Metadata {
	return NewMetadata(Version, Commit, Date, Dirty)
}

func NewMetadata(version string, commit string, date string, dirty string) Metadata {
	return Metadata{
		Version: withDefault(version, defaultVersion),
		Commit:  withDefault(commit, defaultCommit),
		Date:    withDefault(date, defaultDate),
		Dirty:   withDefault(dirty, defaultDirty),
	}
}

func (metadata Metadata) Format() string {
	return fmt.Sprintf(
		"SpecHarbor %s\ncommit: %s\ndate: %s\ndirty: %s",
		metadata.Version,
		metadata.Commit,
		metadata.Date,
		metadata.Dirty,
	)
}

func withDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
