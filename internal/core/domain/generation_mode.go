package domain

type GenerationMode string

const (
	BlankMode         GenerationMode = "blank"
	GuidedMode        GenerationMode = "guided"
	TemplateMode      GenerationMode = "template"
	AIAssistedMode    GenerationMode = "ai-assisted"
	AgentAssistedMode GenerationMode = "agent-assisted"
	HybridMode        GenerationMode = "hybrid"
)
