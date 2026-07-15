package ai

type GenerateInput struct {
	SystemPrompt   string
	UserPrompt     string
	Model          string
	Temperature    float64
	MaxTokens      int
	ResponseSchema map[string]any // optional: only include if structured output schema is needed
}

type GenerateResult struct {
	Text string
}
