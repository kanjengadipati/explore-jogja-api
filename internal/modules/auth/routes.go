package auth

import (
	"pleco-api/internal/config"
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *AuthHandler, jwtService *services.JWTService, rateStore middleware.RateLimitStore, tokenVersionSrc middleware.AccessTokenVersionSource, cfg config.AppConfig) {
	auth := api.Group("/auth")
	if rateStore == nil {
		rateStore = middleware.NewInMemoryRateLimitStore()
	}
	loginRequests := cfg.AuthRateLimit.LoginRequests
	if loginRequests < 1 {
		loginRequests = 5
	}
	loginWindow := time.Duration(cfg.AuthRateLimit.LoginWindowSeconds) * time.Second
	if loginWindow < time.Second {
		loginWindow = 15 * time.Minute
	}
	loginLimiter := middleware.NewRateLimiterWithStore(loginRequests, loginWindow, rateStore)

	registerRequests := cfg.AuthRateLimit.RegisterRequests
	if registerRequests < 1 {
		registerRequests = 3
	}
	registerWindow := time.Duration(cfg.AuthRateLimit.RegisterWindowSeconds) * time.Second
	if registerWindow < time.Second {
		registerWindow = time.Hour
	}
	registerLimiter := middleware.NewRateLimiterWithStore(registerRequests, registerWindow, rateStore)

	passwordLimiter := middleware.NewRateLimiterWithStore(3, time.Hour, rateStore)
	refreshLimiter := middleware.NewRateLimiterWithStore(10, time.Minute, rateStore)
	socialLimiter := middleware.NewRateLimiterWithStore(5, time.Minute, rateStore)
	protectedLimiter := middleware.NewRateLimiterWithStore(60, time.Minute, rateStore)
	otpRequests := cfg.OTPRateLimit.Requests
	if otpRequests < 1 {
		otpRequests = 5
	}
	otpWindow := time.Duration(cfg.OTPRateLimit.WindowSeconds) * time.Second
	if otpWindow < time.Second {
		otpWindow = time.Hour
	}
	otpLimiter := middleware.NewRateLimiterWithStore(otpRequests, otpWindow, rateStore)

	auth.POST("/register", registerLimiter.Middleware(), handler.Register)
	auth.POST("/login", loginLimiter.Middleware(), handler.Login)
	auth.POST("/passwordless/check", otpLimiter.Middleware(), handler.CheckPasswordlessIdentity)
	auth.POST("/passwordless/start", otpLimiter.Middleware(), handler.StartPasswordless)
	auth.POST("/magic-link/verify", loginLimiter.Middleware(), handler.VerifyMagicLink)
	auth.POST("/request-otp", otpLimiter.Middleware(), handler.RequestOTP)
	auth.POST("/verify-otp", loginLimiter.Middleware(), handler.VerifyOTP)
	auth.POST("/refresh", refreshLimiter.Middleware(), handler.RefreshToken)
	auth.GET("/verify", passwordLimiter.Middleware(), handler.VerifyEmail)
	auth.POST("/resend-verification", passwordLimiter.Middleware(), handler.ResendVerification)
	auth.POST("/forgot-password", passwordLimiter.Middleware(), handler.ForgotPassword)
	auth.POST("/reset-password", passwordLimiter.Middleware(), handler.ResetPassword)
	auth.POST("/social-login", socialLimiter.Middleware(), handler.SocialLogin)

	protected := auth.Group("/")
	protected.Use(protectedLimiter.Middleware())
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	protected.GET("/profile", handler.Profile)
	protected.GET("/social/:provider/account", handler.SocialAccount)
	protected.GET("/sessions", handler.ListSessions)
	protected.POST("/logout", handler.Logout)
	protected.POST("/logout-all", handler.LogoutAll)
	protected.POST("/logout-others", handler.LogoutOtherSessions)
	protected.DELETE("/sessions/:id", handler.RevokeSession)
	protected.DELETE("/trusted-devices/:id", handler.RevokeTrustedDevice)
}
