package business

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Business, error)
	FindByID(externalID string) (*Business, error)
	FindByExternalIDs(ids []string) ([]Business, error)
	FindByLegacyPartner(legacyExternalID string) (*Business, error)
	SoftDeleteByLegacyPartner(legacyExternalID string) error
	FindByOwner(userID uint) ([]Business, error)
	FindByIDAndOwner(externalID string, userID uint) (*Business, error)
	FindByStatus(status string) ([]Business, error)
	FindOwnerUserIDs(businessID uint) ([]uint, error)
	ListOwners(businessID uint) ([]BusinessOwner, error)
	FindSimilarByName(query string, status string) ([]Business, error)
	GetListings(businessID uint) ([]OwnedListing, error)
	Create(b *Business) error
	CreateWithServiceAreas(b *Business, regions []string) error
	Update(b *Business) error
	SoftDelete(externalID string) error
	UpsertOwner(businessID, userID uint, role string) error
	IsOwner(businessID, userID uint) (bool, error)
	GetRole(businessID, userID uint) (string, error)
	IsOwnerRole(businessID, userID uint) (bool, error)
	SetInvitedBy(businessID, userID, inviterID uint) error
	RemoveOwner(businessID, userID uint) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Business, error) {
	var businesses []Business
	err := r.db.Order("id ASC").Find(&businesses).Error
	return businesses, err
}

func (r *GormRepository) FindByID(externalID string) (*Business, error) {
	var b Business
	err := r.db.Where("external_id = ?", externalID).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GormRepository) FindByExternalIDs(ids []string) ([]Business, error) {
	var businesses []Business
	err := r.db.Where("external_id IN ?", ids).Find(&businesses).Error
	return businesses, err
}

func (r *GormRepository) FindByLegacyPartner(legacyExternalID string) (*Business, error) {
	var b Business
	err := r.db.Where("legacy_partner_external_id = ?", legacyExternalID).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GormRepository) FindByOwner(userID uint) ([]Business, error) {
	var businesses []Business
	err := r.db.Joins("JOIN business_owners bo ON bo.business_id = businesses.id").
		Where("bo.user_id = ?", userID).
		Order("businesses.id ASC").
		Find(&businesses).Error
	return businesses, err
}

func (r *GormRepository) FindByIDAndOwner(externalID string, userID uint) (*Business, error) {
	var b Business
	err := r.db.Joins("JOIN business_owners bo ON bo.business_id = businesses.id").
		Where("businesses.external_id = ? AND bo.user_id = ?", externalID, userID).
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GormRepository) FindByStatus(status string) ([]Business, error) {
	var businesses []Business
	err := r.db.Where("status = ?", status).Order("id ASC").Find(&businesses).Error
	return businesses, err
}

func (r *GormRepository) FindOwnerUserIDs(businessID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Table("business_owners").
		Where("business_id = ?", businessID).
		Pluck("user_id", &ids).Error
	return ids, err
}

func (r *GormRepository) ListOwners(businessID uint) ([]BusinessOwner, error) {
	var owners []BusinessOwner
	err := r.db.Where("business_id = ?", businessID).
		Order("created_at ASC").
		Find(&owners).Error
	return owners, err
}

// GetListings aggregates the listing rows claimed by a business (business_id FK
// set via listing claims) across the 7 claimable tables. Only destinations and
// events carry a status column; the rest scan as empty.
func (r *GormRepository) GetListings(businessID uint) ([]OwnedListing, error) {
	var listings []OwnedListing
	err := r.db.Raw(`
		SELECT listing_type, external_id, name, status
		FROM (
			SELECT 'destination' AS listing_type, external_id, name, status FROM destinations WHERE business_id = ?
			UNION ALL
			SELECT 'hotel', external_id, name, NULL AS status FROM hotels WHERE business_id = ?
			UNION ALL
			SELECT 'restaurant', external_id, name, NULL AS status FROM restaurants WHERE business_id = ?
			UNION ALL
			SELECT 'souvenir', external_id, name, NULL AS status FROM souvenirs WHERE business_id = ?
			UNION ALL
			SELECT 'rental', external_id, name, NULL AS status FROM rentals WHERE business_id = ?
			UNION ALL
			SELECT 'guide', external_id, name, NULL AS status FROM guides WHERE business_id = ?
			UNION ALL
			SELECT 'event', external_id, title AS name, status FROM events WHERE business_id = ?
		) t
		ORDER BY listing_type, name
	`, businessID, businessID, businessID, businessID, businessID, businessID, businessID).Scan(&listings).Error
	return listings, err
}

func (r *GormRepository) Create(b *Business) error {
	if err := r.db.Create(b).Error; err != nil {
		return err
	}
	subExtID := "sub_" + b.ExternalID
	_ = r.db.Exec(`
		INSERT INTO subscriptions (external_id, business_id, plan, status, created_at, updated_at)
		SELECT ?, ?, 'free', 'active', NOW(), NOW()
		WHERE NOT EXISTS (SELECT 1 FROM subscriptions WHERE business_id = ?)
	`, subExtID, b.ID, b.ID).Error
	return nil
}

// CreateWithServiceAreas creates a business and its service area rows in a
// single transaction, then auto-creates the free subscription.
func (r *GormRepository) CreateWithServiceAreas(b *Business, regions []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		if len(regions) > 0 {
			areas := make([]BusinessServiceArea, len(regions))
			for i, region := range regions {
				areas[i] = BusinessServiceArea{BusinessID: b.ID, Region: region}
			}
			if err := tx.Create(&areas).Error; err != nil {
				return err
			}
		}
		subExtID := "sub_" + b.ExternalID
		return tx.Exec(`
			INSERT INTO subscriptions (external_id, business_id, plan, status, created_at, updated_at)
			SELECT ?, ?, 'free', 'active', NOW(), NOW()
			WHERE NOT EXISTS (SELECT 1 FROM subscriptions WHERE business_id = ?)
		`, subExtID, b.ID, b.ID).Error
	})
}

// FindSimilarByName returns businesses with names matching the query
// (case-insensitive), filtered by status. Used for global name-dedup at
// registration time.
func (r *GormRepository) FindSimilarByName(query string, status string) ([]Business, error) {
	var businesses []Business
	err := r.db.Where("name ILIKE ? AND status = ?", "%"+query+"%", status).
		Limit(5).
		Find(&businesses).Error
	return businesses, err
}

func (r *GormRepository) Update(b *Business) error {
	return r.db.Save(b).Error
}

func (r *GormRepository) SoftDelete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Business{}).Error
}

func (r *GormRepository) SoftDeleteByLegacyPartner(legacyExternalID string) error {
	return r.db.Where("legacy_partner_external_id = ?", legacyExternalID).Delete(&Business{}).Error
}

func (r *GormRepository) UpsertOwner(businessID, userID uint, role string) error {
	return r.db.Exec(`
		INSERT INTO business_owners (business_id, user_id, role, created_at)
		VALUES (?, ?, ?, NOW())
		ON CONFLICT (business_id, user_id)
		DO UPDATE SET role = EXCLUDED.role
	`, businessID, userID, role).Error
}

func (r *GormRepository) IsOwner(businessID, userID uint) (bool, error) {
	var count int64
	err := r.db.Table("business_owners").
		Where("business_id = ? AND user_id = ?", businessID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *GormRepository) GetRole(businessID, userID uint) (string, error) {
	var role string
	err := r.db.Table("business_owners").
		Where("business_id = ? AND user_id = ?", businessID, userID).
		Pluck("role", &role).Error
	return role, err
}

func (r *GormRepository) IsOwnerRole(businessID, userID uint) (bool, error) {
	var count int64
	err := r.db.Table("business_owners").
		Where("business_id = ? AND user_id = ? AND role = ?", businessID, userID, RoleOwner).
		Count(&count).Error
	return count > 0, err
}

func (r *GormRepository) RemoveOwner(businessID, userID uint) error {
	return r.db.Where("business_id = ? AND user_id = ?", businessID, userID).
		Delete(&BusinessOwner{}).Error
}

func (r *GormRepository) SetInvitedBy(businessID, userID, inviterID uint) error {
	return r.db.Exec(`
		UPDATE business_owners SET invited_by = ?
		WHERE business_id = ? AND user_id = ?
	`, inviterID, businessID, userID).Error
}
