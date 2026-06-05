package errors

import "errors"

var (
	ErrTelegramNotificationNotFound       = errors.New("telegram notification not found")
	ErrInvalidTelegramNotificationType    = errors.New("invalid telegram notification type")
	ErrInvalidTelegramNotificationStatus  = errors.New("invalid telegram notification status")
	ErrInvalidTelegramNotificationPayload = errors.New("invalid telegram notification payload")
	ErrInvalidTelegramNotificationEntity  = errors.New("invalid telegram notification entity state")
	ErrTelegramNotificationNotRetryable   = errors.New("only failed telegram notifications can be retried")
)
