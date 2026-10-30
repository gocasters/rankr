package delivery

import "errors"

var (
	ErrWebhookHmacSigFailure = errors.New("webhook hmac signature verification failed")
)
