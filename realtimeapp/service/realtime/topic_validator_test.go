package realtime

import (
	"testing"

	"github.com/gocasters/rankr/pkg/topicsname"
	"github.com/stretchr/testify/assert"
)

func TestTopicValidator_ValidateTopics(t *testing.T) {
	validator := NewTopicValidator()

	tests := []struct {
		name             string
		topics           []string
		clientPerms      ClientPermissions
		expectedAllowed  []string
		expectedDenied   map[string]string
		checkDeniedCount bool
	}{
		{
			name: "all public topics allowed for authenticated client",
			topics: []string{
				topicsname.TopicContributorCreated,
				topicsname.TopicTaskUpdated,
				topicsname.TopicLeaderboardScored,
			},
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectedAllowed: []string{
				topicsname.TopicContributorCreated,
				topicsname.TopicTaskUpdated,
				topicsname.TopicLeaderboardScored,
			},
			expectedDenied: map[string]string{},
		},
		{
			name: "empty topic denied",
			topics: []string{
				topicsname.TopicContributorCreated,
				"",
				topicsname.TopicTaskUpdated,
			},
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectedAllowed: []string{
				topicsname.TopicContributorCreated,
				topicsname.TopicTaskUpdated,
			},
			checkDeniedCount: true,
		},
		{
			name: "unknown topic denied",
			topics: []string{
				topicsname.TopicContributorCreated,
				"unknown.topic",
				"malicious.admin.topic",
			},
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectedAllowed: []string{
				topicsname.TopicContributorCreated,
			},
			checkDeniedCount: true,
		},
		{
			name: "whitespace-only topic denied",
			topics: []string{
				topicsname.TopicTaskCreated,
				"   ",
				topicsname.TopicProjectCreated,
			},
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectedAllowed: []string{
				topicsname.TopicTaskCreated,
				topicsname.TopicProjectCreated,
			},
			checkDeniedCount: true,
		},
		{
			name: "all valid public topics",
			topics: []string{
				topicsname.TopicContributorCreated,
				topicsname.TopicContributorUpdated,
				topicsname.TopicTaskCreated,
				topicsname.TopicTaskUpdated,
				topicsname.TopicTaskCompleted,
				topicsname.TopicLeaderboardScored,
				topicsname.TopicLeaderboardUpdated,
				topicsname.TopicProjectCreated,
				topicsname.TopicProjectUpdated,
			},
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectedAllowed: []string{
				topicsname.TopicContributorCreated,
				topicsname.TopicContributorUpdated,
				topicsname.TopicTaskCreated,
				topicsname.TopicTaskUpdated,
				topicsname.TopicTaskCompleted,
				topicsname.TopicLeaderboardScored,
				topicsname.TopicLeaderboardUpdated,
				topicsname.TopicProjectCreated,
				topicsname.TopicProjectUpdated,
			},
			expectedDenied: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, denied := validator.ValidateTopics(tt.topics, tt.clientPerms)

			assert.Equal(t, tt.expectedAllowed, allowed, "allowed topics mismatch")

			if tt.checkDeniedCount {
				assert.NotEmpty(t, denied, "expected some topics to be denied")
			} else {
				assert.Equal(t, tt.expectedDenied, denied, "denied topics mismatch")
			}
		})
	}
}

func TestTopicValidator_validateTopic(t *testing.T) {
	validator := NewTopicValidator()

	tests := []struct {
		name        string
		topic       string
		clientPerms ClientPermissions
		expectError bool
		errorMsg    string
	}{
		{
			name:  "valid public topic",
			topic: topicsname.TopicContributorCreated,
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectError: false,
		},
		{
			name:  "empty topic",
			topic: "",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectError: true,
			errorMsg:    "topic cannot be empty",
		},
		{
			name:  "whitespace topic",
			topic: "   ",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectError: true,
			errorMsg:    "topic cannot be empty",
		},
		{
			name:  "unknown topic not in whitelist",
			topic: "unknown.topic",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectError: true,
			errorMsg:    "topic not allowed: unknown.topic",
		},
		{
			name:  "malicious topic attempt",
			topic: "admin.internal.secret",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
			},
			expectError: true,
			errorMsg:    "topic not allowed: admin.internal.secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateTopic(tt.topic, tt.clientPerms)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultClientPermissions(t *testing.T) {
	perms := DefaultClientPermissions()

	assert.True(t, perms.IsAuthenticated, "should be authenticated by default")
	assert.False(t, perms.CanAccessPrivateTopics, "should not have private topic access by default")
	assert.Empty(t, perms.UserID, "user ID should be empty by default")
	assert.Empty(t, perms.Roles, "roles should be empty by default")
}

func TestNewTopicValidator(t *testing.T) {
	validator := NewTopicValidator()

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.topicRules)
	assert.NotNil(t, validator.topicPatterns)

	// Verify all expected public topics are registered
	expectedPublicTopics := []string{
		topicsname.TopicContributorCreated,
		topicsname.TopicContributorUpdated,
		topicsname.TopicTaskCreated,
		topicsname.TopicTaskUpdated,
		topicsname.TopicTaskCompleted,
		topicsname.TopicLeaderboardScored,
		topicsname.TopicLeaderboardUpdated,
		topicsname.TopicProjectCreated,
		topicsname.TopicProjectUpdated,
	}

	for _, topic := range expectedPublicTopics {
		accessLevel, exists := validator.topicRules[topic]
		assert.True(t, exists, "topic %s should be registered", topic)
		assert.Equal(t, TopicAccessPublic, accessLevel, "topic %s should be public", topic)
	}
}

func TestTopicValidator_PatternMatching(t *testing.T) {
	validator := &TopicValidator{
		topicRules: map[string]TopicAccessLevel{
			"task.created": TopicAccessPublic,
		},
		topicPatterns: map[string]TopicAccessLevel{
			"user.*.read":          TopicAccessPrivate,
			"user.*.notifications": TopicAccessPrivate,
			"org.*.updates":        TopicAccessPrivate,
			"team.*.chat":          TopicAccessPrivate,
		},
	}

	tests := []struct {
		name             string
		topic            string
		clientPerms      ClientPermissions
		expectError      bool
		errorMsgContains string
	}{
		{
			name:  "pattern match user.123.read with correct userID",
			topic: "user.123.read",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "123",
			},
			expectError: false,
		},
		{
			name:  "pattern match user.123.read with wrong userID",
			topic: "user.123.read",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "456",
			},
			expectError:      true,
			errorMsgContains: "cannot subscribe to other user's topic",
		},
		{
			name:  "pattern match user.456.notifications with correct userID",
			topic: "user.456.notifications",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "456",
			},
			expectError: false,
		},
		{
			name:  "pattern match without userID in permissions",
			topic: "user.123.read",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "",
			},
			expectError:      true,
			errorMsgContains: "user ID not available",
		},
		{
			name:  "pattern match org.acme.updates",
			topic: "org.acme.updates",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "123",
			},
			expectError: false,
		},
		{
			name:  "pattern match team.engineering.chat",
			topic: "team.engineering.chat",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "123",
			},
			expectError: false,
		},
		{
			name:  "no pattern match for unknown topic",
			topic: "unknown.topic.format",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: true,
				UserID:                 "123",
			},
			expectError:      true,
			errorMsgContains: "topic not allowed",
		},
		{
			name:  "exact match takes precedence over pattern",
			topic: "task.created",
			clientPerms: ClientPermissions{
				IsAuthenticated:        true,
				CanAccessPrivateTopics: false,
				UserID:                 "123",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateTopic(tt.topic, tt.clientPerms)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsgContains != "" {
					assert.Contains(t, err.Error(), tt.errorMsgContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTopicValidator_MatchPattern(t *testing.T) {
	validator := &TopicValidator{
		topicPatterns: map[string]TopicAccessLevel{
			"user.*.read":          TopicAccessPrivate,
			"user.*.notifications": TopicAccessPrivate,
			"org.*.updates":        TopicAccessPublic,
			"team.*.*.chat":        TopicAccessPrivate, // multi-level pattern
		},
	}

	tests := []struct {
		name          string
		topic         string
		expectPattern string
		expectAccess  TopicAccessLevel
		shouldMatch   bool
	}{
		{
			name:          "match user.123.read",
			topic:         "user.123.read",
			expectPattern: "user.*.read",
			expectAccess:  TopicAccessPrivate,
			shouldMatch:   true,
		},
		{
			name:          "match user.abc.notifications",
			topic:         "user.abc.notifications",
			expectPattern: "user.*.notifications",
			expectAccess:  TopicAccessPrivate,
			shouldMatch:   true,
		},
		{
			name:          "match org.acme.updates",
			topic:         "org.acme.updates",
			expectPattern: "org.*.updates",
			expectAccess:  TopicAccessPublic,
			shouldMatch:   true,
		},
		{
			name:        "no match for user.123.write",
			topic:       "user.123.write",
			shouldMatch: false,
		},
		{
			name:        "no match for wrong depth",
			topic:       "user.read",
			shouldMatch: false,
		},
		{
			name:          "match multi-level pattern",
			topic:         "team.eng.dev.chat",
			expectPattern: "team.*.*.chat",
			expectAccess:  TopicAccessPrivate,
			shouldMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, accessLevel := validator.matchPattern(tt.topic)

			if tt.shouldMatch {
				assert.Equal(t, tt.expectPattern, pattern)
				assert.Equal(t, tt.expectAccess, accessLevel)
			} else {
				assert.Empty(t, pattern)
			}
		})
	}
}

func TestTopicValidator_SameUserMultipleSubscriptions(t *testing.T) {
	validator := &TopicValidator{
		topicRules: map[string]TopicAccessLevel{
			"task.created": TopicAccessPublic,
		},
		topicPatterns: map[string]TopicAccessLevel{
			"user.*.read":          TopicAccessPrivate,
			"user.*.notifications": TopicAccessPrivate,
		},
	}

	clientPerms := ClientPermissions{
		IsAuthenticated:        true,
		CanAccessPrivateTopics: true,
		UserID:                 "123",
	}

	// Test subscribing to the same topic multiple times
	topics := []string{
		"user.123.read",
		"user.123.read", // duplicate - should be allowed
		"user.123.notifications",
		"user.123.read", // duplicate again - should be allowed
	}

	allowed, denied := validator.ValidateTopics(topics, clientPerms)

	// All should be allowed (validation doesn't prevent duplicates)
	assert.Equal(t, 4, len(allowed), "all topics should be allowed including duplicates")
	assert.Empty(t, denied, "no topics should be denied")

	// Verify each duplicate is in the allowed list
	assert.Contains(t, allowed, "user.123.read")
	assert.Contains(t, allowed, "user.123.notifications")

	// Count occurrences of user.123.read
	count := 0
	for _, topic := range allowed {
		if topic == "user.123.read" {
			count++
		}
	}
	assert.Equal(t, 3, count, "user.123.read should appear 3 times in allowed list")
}

func TestTopicValidator_MultipleUsersCannotAccessEachOther(t *testing.T) {
	validator := &TopicValidator{
		topicPatterns: map[string]TopicAccessLevel{
			"user.*.read":          TopicAccessPrivate,
			"user.*.notifications": TopicAccessPrivate,
		},
	}

	// User 123 tries to subscribe to their own and another user's topics
	user123Perms := ClientPermissions{
		IsAuthenticated:        true,
		CanAccessPrivateTopics: true,
		UserID:                 "123",
	}

	topics := []string{
		"user.123.read",          // ✅ own topic
		"user.123.notifications", // ✅ own topic
		"user.456.read",          // ❌ another user's topic
		"user.789.notifications", // ❌ another user's topic
	}

	allowed, denied := validator.ValidateTopics(topics, user123Perms)

	// Should only allow own topics
	assert.Equal(t, 2, len(allowed), "should allow only own topics")
	assert.Equal(t, 2, len(denied), "should deny other users' topics")

	assert.Contains(t, allowed, "user.123.read")
	assert.Contains(t, allowed, "user.123.notifications")

	assert.Contains(t, denied, "user.456.read")
	assert.Contains(t, denied, "user.789.notifications")
	assert.Contains(t, denied["user.456.read"], "cannot subscribe to other user's topic")
	assert.Contains(t, denied["user.789.notifications"], "cannot subscribe to other user's topic")
}
