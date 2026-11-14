package delivery

import "errors"

var (
	ErrWebhookHmacSigFailure = errors.New("webhook hmac signature verification failed")
	// ErrWebhookConfig marks server-side misconfiguration; callers must not
	// expose its detail to unauthenticated clients.
	ErrWebhookConfig = errors.New("webhook service misconfigured")
)
