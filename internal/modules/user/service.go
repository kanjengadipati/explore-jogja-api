package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"pleco-api/internal/cache"
	"pleco-api/internal/domain"
	"pleco-api/internal/modules/audit"
	roleModule "pleco-api/internal/modules/role"
	tokenModule "pleco-api/internal/modules/token"
	"pleco-api/internal/services"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var errNotSalesUser = errors.New("only users with the sales role have a referral code")

var errUnsupportedRole = errors.New("unsupported role")

type Service struct {
	DB               *gorm.DB
	UserRepo         Repository
	RefreshTokenRepo tokenModule.RefreshTokenRepository
	AuditSvc         *audit.Service
	Cache            cache.Store
}

func NewService(db *gorm.DB, userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository, auditSvc *audit.Service) *Service {
	return &Service{DB: db, UserRepo: userRepo, RefreshTokenRepo: refreshRepo, AuditSvc: auditSvc}
}

func (s *Service) GetAllUsers(page, limit int, search, role string) ([]User, int64, error) {
	return s.UserRepo.FindAllWithFilter(page, limit, search, role)
}

func (s *Service) GetUserByID(id uint) (*User, error) {
	if s.Cache != nil {
		var user User
		key := fmt.Sprintf("user:detail:%d", id)
		if ok, err := s.Cache.GetJSON(context.Background(), key, &user); err == nil && ok {
			return &user, nil
		}
		found, err := s.UserRepo.FindByID(id)
		if err != nil {
			return nil, err
		}
		_ = s.Cache.SetJSON(context.Background(), key, found, 5*time.Minute)
		return found, nil
	}

	return s.UserRepo.FindByID(id)
}

func (s *Service) CreateUser(input CreateUserRequest) (*User, error) {
	if existing, err := s.UserRepo.FindByEmail(input.Email); err == nil && existing != nil && existing.ID != 0 {
		return nil, domain.NewAPIError(http.StatusConflict, domain.CodeConflict, "email already in use", domain.ErrConflict)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := input.Role
	if role == "" {
		role = "user"
	}
	if !isAssignableRole(role) {
		return nil, errUnsupportedRole
	}

	hashedPassword, err := services.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		Name:               input.Name,
		Email:              input.Email,
		Password:           hashedPassword,
		Role:               role,
		IsVerified:         input.IsVerified,
		PasswordUpdatedAt:  now,
		LastPasswordChange: ptrTime(now),
	}

	if err := s.UserRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateUser(id uint, input UpdateUserRequest) (*User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !isAssignableRole(input.Role) {
		return nil, errUnsupportedRole
	}

	if existing, err := s.UserRepo.FindByEmail(input.Email); err == nil && existing != nil && existing.ID != 0 && existing.ID != id {
		return nil, errors.New("email already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	oldRole := user.Role
	user.Name = input.Name
	user.Email = input.Email
	user.Role = input.Role
	user.IsVerified = input.IsVerified

	if oldRole != user.Role {
		user.AccessTokenVersion++
	}

	if oldRole != user.Role {
		if err := s.runInTx(func(userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository) error {
			if err := userRepo.Update(user); err != nil {
				return err
			}
			return refreshRepo.DeleteByUser(user.ID)
		}); err != nil {
			return nil, err
		}
		s.invalidateUserCache(user.ID)
		return user, nil
	}

	if err := s.UserRepo.Update(user); err != nil {
		return nil, err
	}
	s.invalidateUserCache(user.ID)
	return user, nil
}

func (s *Service) UpdateProfile(id uint, input UpdateProfileRequest) (*User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	user.Name = input.Name
	user.PhoneNumber = input.PhoneNumber
	// Only overwrite avatar/cover when the caller explicitly provides a value,
	// preventing partial updates (e.g. avatar-only) from blanking the other field.
	if input.AvatarURL != "" {
		user.AvatarURL = input.AvatarURL
	}
	if input.CoverImageURL != "" {
		user.CoverImageURL = input.CoverImageURL
	}
	if err := s.UserRepo.Update(user); err != nil {
		return nil, err
	}
	s.invalidateUserCache(user.ID)

	return user, nil
}

func (s *Service) ChangePassword(id uint, currentPassword, newPassword string) error {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := services.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	now := time.Now()
	user.PasswordUpdatedAt = now
	user.LastPasswordChange = &now
	user.AccessTokenVersion++

	if err := s.runInTx(func(userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository) error {
		if err := userRepo.Update(user); err != nil {
			return err
		}
		return refreshRepo.DeleteByUser(user.ID)
	}); err != nil {
		return err
	}
	s.invalidateUserCache(user.ID)
	return nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

// PromoteToPartnerRole upgrades a user's role string to "partner" and
// bumps their access token version so old tokens are invalidated.
// The caller (business.Service) is responsible for running this
// inside the same transaction that creates the business listing.
func (s *Service) PromoteToPartnerRole(userID uint) error {
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.Role == "partner" {
		return nil
	}
	if user.Role == "admin" || user.Role == "superadmin" {
		return errors.New("cannot downgrade elevated role to partner")
	}

	user.Role = "partner"
	// Keep role_id in sync with the role string so both sources agree.
	var partnerRole roleModule.Role
	if err := s.DB.Where("name = ?", "partner").First(&partnerRole).Error; err == nil {
		user.RoleID = partnerRole.ID
	}
	// Do NOT bump AccessTokenVersion here — partner promotion is an upgrade,
	// not a security event. Existing refresh tokens stay valid so the user
	// can get a new access token (with the 'partner' role) via a normal refresh.

	if err := s.UserRepo.Update(user); err != nil {
		return err
	}
	s.invalidateUserCache(user.ID)
	return nil
}

// --- Sales referral (used by the commission feature) ---

// GetOrCreateReferralCode returns the sales user's referral code, generating
// one on first use. Only users with Role == "sales" have a code.
func (s *Service) GetOrCreateReferralCode(userID uint) (string, error) {
	u, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	if u.Role != "sales" {
		return "", errNotSalesUser
	}
	if u.ReferralCode != nil && *u.ReferralCode != "" {
		return *u.ReferralCode, nil
	}
	return s.assignFreshReferralCode(u)
}

// RegenerateReferralCode issues a new code, invalidating the old one for
// future signups. Partners already linked keep their ReferredBySalesID.
func (s *Service) RegenerateReferralCode(userID uint) (string, error) {
	u, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	if u.Role != "sales" {
		return "", errNotSalesUser
	}
	return s.assignFreshReferralCode(u)
}

func (s *Service) assignFreshReferralCode(u *User) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		code := generateReferralCode(u.Name, u.ID, attempt)
		u.ReferralCode = &code
		if err := s.UserRepo.Update(u); err != nil {
			continue // likely a rare unique-index clash on the suffix — retry
		}
		return code, nil
	}
	return "", errors.New("failed to generate a unique referral code, please try again")
}

// GetIDByReferralCode resolves a referral code to its owning sales user's ID.
// Used by business.Service when a partner signs up through a referral link.
func (s *Service) GetIDByReferralCode(code string) (uint, error) {
	var u User
	if err := s.DB.Where("referral_code = ? AND role = ?", code, "sales").First(&u).Error; err != nil {
		return 0, err
	}
	return u.ID, nil
}

// SetReferredBySales stamps a user (partner) with the sales user who referred
// them. No-op if already set — first referral wins, later signup attempts
// with a different code never overwrite it.
func (s *Service) SetReferredBySales(userID, salesUserID uint) error {
	u, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if u.ReferredBySalesID != nil {
		return nil
	}
	u.ReferredBySalesID = &salesUserID
	return s.UserRepo.Update(u)
}

func generateReferralCode(name string, userID uint, attempt int) string {
	base := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), " ", ""))
	if len(base) > 6 {
		base = base[:6]
	}
	if base == "" {
		base = "SALES"
	}
	if attempt == 0 {
		return fmt.Sprintf("%s%d", base, userID)
	}
	return fmt.Sprintf("%s%d%d", base, userID, attempt)
}

func (s *Service) DeleteUser(id uint, callerRole string, callerID uint) error {
	if id == callerID {
		return errors.New("cannot delete yourself")
	}

	targetUser, err := s.UserRepo.FindByID(id)
	if err != nil {
		return err
	}

	if callerRole == "admin" && targetUser.Role != "user" && targetUser.Role != "partner" {
		return errors.New("admin can only delete standard users or partners")
	}

	if targetUser.Role == "superadmin" {
		return errors.New("cannot delete superadmin")
	}

	if err := s.UserRepo.Delete(id); err != nil {
		return err
	}
	s.invalidateUserCache(id)
	return nil
}

func (s *Service) invalidateUserCache(userID uint) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.Delete(
		context.Background(),
		fmt.Sprintf("user:detail:%d", userID),
		fmt.Sprintf("user:profile:%d", userID),
		fmt.Sprintf("user:permissions:%d", userID),
	)
}

func (s *Service) runInTx(fn func(userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.RefreshTokenRepo)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.RefreshTokenRepo.WithTx(tx))
	})
}

func isAssignableRole(role string) bool {
	switch role {
	case "admin", "user", "partner", "sales":
		return true
	default:
		return false
	}
}
