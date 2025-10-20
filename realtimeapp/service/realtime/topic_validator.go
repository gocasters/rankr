package realtime

import (
	"fmt"
	"strings"

	"github.com/gocasters/rankr/pkg/topicsname"
)

type TopicAccessLevel string

const (
	TopicAccessPublic  TopicAccessLevel = "public"
	TopicAccessPrivate TopicAccessLevel = "private"
)

type TopicValidator struct {
	topicRules map[string]TopicAccessLevel

	topicPatterns map[string]TopicAccessLevel
}

// NewTopicValidator creates a new topic validator with default rules
func NewTopicValidator() *TopicValidator {
	return &TopicValidator{
		topicRules: map[string]TopicAccessLevel{
			topicsname.TopicContributorCreated: TopicAccessPublic,
			topicsname.TopicContributorUpdated: TopicAccessPublic,
			topicsname.TopicTaskCreated:        TopicAccessPublic,
			topicsname.TopicTaskUpdated:        TopicAccessPublic,
			topicsname.TopicTaskCompleted:      TopicAccessPublic,
			topicsname.TopicLeaderboardScored:  TopicAccessPublic,
			topicsname.TopicLeaderboardUpdated: TopicAccessPublic,
			topicsname.TopicProjectCreated:     TopicAccessPublic,
			topicsname.TopicProjectUpdated:     TopicAccessPublic,

			//  Example Private Topics:
			// "admin.user.banned":      TopicAccessPrivate,
			// "admin.system.alert":     TopicAccessPrivate,
			// "admin.metrics.internal": TopicAccessPrivate,
		},
		topicPatterns: map[string]TopicAccessLevel{
			// "user.*.notifications": TopicAccessPrivate,
			// "project.*.private":    TopicAccessPrivate,
		},
	}
}

func (v *TopicValidator) ValidateTopics(topics []string, clientPerms ClientPermissions) (allowed []string, err error) {
	allowed = make([]string, 0, len(topics))
	hasInvalidTopics := false

	for _, topic := range topics {
		if err := v.validateTopic(topic, clientPerms); err != nil {
			hasInvalidTopics = true
		} else {
			allowed = append(allowed, topic)
		}
	}

	if hasInvalidTopics {
		err = fmt.Errorf("some topics are not allowed")
	}

	return allowed, err
}

func (v *TopicValidator) validateTopic(topic string, clientPerms ClientPermissions) error {
	if strings.TrimSpace(topic) == "" {
		return fmt.Errorf("topic cannot be empty")
	}

	if accessLevel, exists := v.topicRules[topic]; exists {
		return v.checkAccess(topic, accessLevel, clientPerms)
	}

	if pattern, accessLevel := v.matchPattern(topic); pattern != "" {
		if accessLevel == TopicAccessPrivate {
			if err := v.validateResourceAccess(topic, pattern, clientPerms); err != nil {
				return err
			}
		}
		return v.checkAccess(topic, accessLevel, clientPerms)
	}

	return fmt.Errorf("topic not allowed: %s", topic)
}

func (v *TopicValidator) matchPattern(topic string) (string, TopicAccessLevel) {
	parts := strings.Split(topic, ".")

	for pattern, accessLevel := range v.topicPatterns {
		patternParts := strings.Split(pattern, ".")

		if len(parts) != len(patternParts) {
			continue
		}

		match := true
		for i := range patternParts {
			if patternParts[i] != "*" && patternParts[i] != parts[i] {
				match = false
				break
			}
		}

		if match {
			return pattern, accessLevel
		}
	}

	return "", ""
}

func (v *TopicValidator) validateResourceAccess(topic, pattern string, clientPerms ClientPermissions) error {
	topicParts := strings.Split(topic, ".")
	patternParts := strings.Split(pattern, ".")

	for i, patternPart := range patternParts {
		if patternPart == "*" && i < len(topicParts) {
			resourceID := topicParts[i]

			if i > 0 && patternParts[0] == "user" {
				if clientPerms.UserID == "" {
					return fmt.Errorf("user ID not available in permissions")
				}
				if clientPerms.UserID != resourceID {
					return fmt.Errorf("cannot subscribe to other user's topic: %s", topic)
				}
			}
		}
	}

	return nil
}

func (v *TopicValidator) checkAccess(topic string, accessLevel TopicAccessLevel, clientPerms ClientPermissions) error {
	switch accessLevel {
	case TopicAccessPublic:
		return nil

	case TopicAccessPrivate:
		if !clientPerms.IsAuthenticated {
			return fmt.Errorf("authentication required for topic: %s", topic)
		}
		if !clientPerms.CanAccessPrivateTopics {
			return fmt.Errorf("insufficient permissions for topic: %s", topic)
		}
		return nil

	default:
		return fmt.Errorf("unknown access level for topic: %s", topic)
	}
}

type ClientPermissions struct {
	IsAuthenticated        bool
	CanAccessPrivateTopics bool
	UserID                 string
	Roles                  []string
}

// DefaultClientPermissions returns default permissions for unauthenticated clients
// In this implementation, we treat all WebSocket clients as authenticated
// In a production system, you should validate JWT tokens or session cookies
func DefaultClientPermissions() ClientPermissions {
	return ClientPermissions{
		IsAuthenticated:        true, // TODO: Implement proper authentication
		CanAccessPrivateTopics: false,
		UserID:                 "",
		Roles:                  []string{},
	}
}
