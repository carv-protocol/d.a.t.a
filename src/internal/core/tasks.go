package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

// TaskResult represents the result of executing a task
type TaskResult struct {
	TaskID    string
	Task      *Task
	Actions   []actions.IAction
	Timestamp time.Time
	Result    []error
}

// TaskGeneration represents the result of generating tasks
type TaskGeneration struct {
	Chain *ThoughtChain
	Tasks []*Task
}

// convertThoughtChainToTasks converts a thought chain into concrete tasks
func convertThoughtChainToTasks(chain *ThoughtChain) (*Task, error) {
	if len(chain.Steps) == 0 {
		return nil, fmt.Errorf("thought chain is empty")
	}
	content := chain.Steps[len(chain.Steps)-1].Content
	startTag := "\n```json\n"
	endTag := "\n```\n"

	startIndex := strings.Index(content, startTag)
	if startIndex == -1 {
		return nil, fmt.Errorf("start tag not found")
	}
	startIndex += len(startTag)

	endIndex := strings.Index(content[startIndex:], endTag)
	if endIndex == -1 {
		return nil, fmt.Errorf("end tag not found")
	}
	endIndex += startIndex

	jsonContent := content[startIndex:endIndex]

	var task Task
	if err := json.Unmarshal([]byte(jsonContent), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &task, nil
}
