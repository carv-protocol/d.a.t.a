package tools

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

type TwitterTool struct {
}

func (t *TwitterTool) Initialize(ctx context.Context) error {
	return nil
}

func (t *TwitterTool) Name() string {
	return "twitter tool"
}

func (t *TwitterTool) Description() string {
	return "This is a twitter tool. It allows AI Agent to own a twitter account. It can let AI Agent to post tweets, follow users, like tweets, reply tweets. It can also fetch the tweets of a user, fetch the trending tweets."
}

func (t *TwitterTool) AvailableActions() []actions.IAction {
	return nil
}
