package domain

// TemplateSource identifies where template generation content came from.
type TemplateSource string

const (
	BuiltInTemplateSource TemplateSource = "built-in"
	CustomTemplateSource  TemplateSource = "custom"
)
