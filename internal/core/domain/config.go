package domain

const SupportedLocalConfigVersion = 1

type LocalConfig struct {
	Version    int
	Defaults   ConfigDefaults
	Validation ConfigValidation
	Review     ConfigReview
	Archive    ConfigArchive
	Scan       ConfigScan
	Output     ConfigOutput
}

type ConfigDefaults struct {
	AgentRole      string
	GenerationMode string
}

type ConfigValidation struct {
	RequireAllChangeFiles bool
}

type ConfigReview struct {
	RequireCompletedTasks bool
}

type ConfigArchive struct {
	DateLayout string
}

type ConfigScan struct {
	IncludeCommonProjectFiles bool
}

type ConfigOutput struct {
	Format string
}

type ConfigResult struct {
	Path   string
	Config LocalConfig
}

func IsSupportedLocalConfigVersion(version int) bool {
	return version == SupportedLocalConfigVersion
}
