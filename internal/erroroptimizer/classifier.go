package erroroptimizer

import (
	"errors"
	"strings"

	"pleco-api/internal/domain"
	"pleco-api/internal/services"
)

type ErrorClassifier interface {
	Classify(err error, endpoint string) *ErrorMetadata
}

type DefaultErrorClassifier struct{}

func (dec *DefaultErrorClassifier) Classify(err error, endpoint string) *ErrorMetadata {
	if err == nil {
		return nil
	}
	// Check for known error types
	switch {
	case errors.Is(err, services.ErrWeakPassword):
		return &ErrorMetadata{
			Code:                "AUTH_WEAK_PASSWORD",
			Type:                "validation",
			Severity:            "warning",
			UserMessage:         "Password doesn't meet security requirements",
			ShouldExposeDetails: true,
		}

	case errors.Is(err, domain.ErrInvalidCredentials):
		return &ErrorMetadata{
			Code:                "AUTH_INVALID_CREDENTIALS",
			Type:                "authentication",
			Severity:            "error",
			UserMessage:         "Invalid email or password",
			ShouldExposeDetails: false,
		}

	case errors.Is(err, domain.ErrAccountLocked):
		return &ErrorMetadata{
			Code:                "AUTH_ACCOUNT_LOCKED",
			Type:                "authentication",
			Severity:            "warning",
			UserMessage:         "Your account is locked due to too many failed attempts",
			ShouldExposeDetails: false,
		}

	case strings.Contains(err.Error(), "rate limit") || strings.Contains(strings.ToLower(err.Error()), "too many requests"):
		return &ErrorMetadata{
			Code:                "RATE_LIMIT_EXCEEDED",
			Type:                "rate_limit",
			Severity:            "warning",
			UserMessage:         "Too many attempts. Please try again later.",
			ShouldExposeDetails: false,
		}

	case errors.Is(err, domain.ErrEmailNotVerified):
		return &ErrorMetadata{
			Code:                "AUTH_EMAIL_NOT_VERIFIED",
			Type:                "authorization",
			Severity:            "warning",
			UserMessage:         "Please verify your email before signing in",
			ShouldExposeDetails: false,
		}

	case strings.Contains(strings.ToLower(err.Error()), "database") || strings.Contains(strings.ToLower(err.Error()), "sql"):
		return &ErrorMetadata{
			Code:                "SERVER_DATABASE_ERROR",
			Type:                "server",
			Severity:            "critical",
			UserMessage:         "We're experiencing technical difficulties",
			ShouldExposeDetails: false,
		}

	case strings.Contains(err.Error(), "passwordless account not found"):
		return &ErrorMetadata{
			Code:                "AUTH_USER_NOT_FOUND",
			Type:                "authentication",
			Severity:            "error",
			UserMessage:         "No account found for this email or WhatsApp number.",
			ShouldExposeDetails: false,
		}

	case strings.Contains(err.Error(), "no whatsapp number available for this account"):
		return &ErrorMetadata{
			Code:                "AUTH_WHATSAPP_NOT_LINKED",
			Type:                "validation",
			Severity:            "warning",
			UserMessage:         "No WhatsApp number is available for this account.",
			ShouldExposeDetails: false,
		}

	case strings.Contains(err.Error(), "invalid or expired OTP"):
		return &ErrorMetadata{
			Code:                "AUTH_INVALID_OTP",
			Type:                "authentication",
			Severity:            "error",
			UserMessage:         "Invalid or expired OTP code.",
			ShouldExposeDetails: false,
		}

	default:
		return &ErrorMetadata{
			Code:                "SERVER_INTERNAL_ERROR",
			Type:                "server",
			Severity:            "error",
			UserMessage:         "Something went wrong. Please try again.",
			ShouldExposeDetails: false,
		}
	}
}
