package characters

type Character struct {
	Name             string
	System           string
	Bio              []string
	Lore             []string
	Style            StyleGuide
	Topics           []string
	Goals            []Goal
	MessageExamples  []string
	TaskInstructions string
	PriorityAccounts []Account
	Preferences      map[string]float64
}

type CharacterConfig struct {
	Name             string   `json:"name"`
	System           string   `json:"system"`
	Bio              []string `json:"bio"`
	Lore             []string `json:"lore"`
	MessageExamples  []string `json:"message_examples"`
	TaskInstructions string   `json:"task_instructions"`
	Style            struct {
		Tone        []string `json:"tone"`
		Constraints []string `json:"constraints"`
	} `json:"style"`
	Topics           []string           `json:"topics"`
	Goals            []Goal             `json:"goals"`
	PriorityAccounts []Account          `json:"priority_accounts"`
	Preferences      map[string]float64 `json:"preferences"`
}

type Goal struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Priority    float64 `json:"priority"`
}

type Account struct {
	Platform string
	ID       string
}

type StyleGuide struct {
	Tone        []string
	Constraints []string
}
