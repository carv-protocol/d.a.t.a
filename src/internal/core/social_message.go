package core

import "github.com/carv-protocol/d.a.t.a/src/internal/types"

// SocialMessageWrapper wraps types.SocialMessage to implement additional interfaces
type SocialMessageWrapper struct {
	*types.SocialMessage
}

// Implement pluginCore.SocialMessage interface
func (m *SocialMessageWrapper) GetPlatform() string {
	return m.Platform
}

func (m *SocialMessageWrapper) GetFromUser() string {
	return m.FromUser
}

func (m *SocialMessageWrapper) GetContent() string {
	return m.Content
}

func (m *SocialMessageWrapper) GetMetadata() map[string]interface{} {
	return m.Metadata
}

// NewSocialMessageWrapper creates a new wrapper for types.SocialMessage
func NewSocialMessageWrapper(msg *types.SocialMessage) *SocialMessageWrapper {
	return &SocialMessageWrapper{msg}
}
