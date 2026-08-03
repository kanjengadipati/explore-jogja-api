package listingclaim

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

var (
	ErrClaimNotFound       = errors.New("claim not found")
	ErrClaimAlreadyDone    = errors.New("claim already reviewed")
	ErrListingNotOwnable   = errors.New("listing type is not supported")
	ErrListingAlreadyOwned = errors.New("listing is already claimed by another business")
)

type Repository interface {
	Create(claim *ListingClaim) error
	FindByExternalID(externalID string) (*ListingClaim, error)
	FindPending() ([]ListingClaim, error)
	FindByBusiness(businessID uint) ([]ListingClaim, error)
	Approve(externalID string, adminUserID uint) error
	Reject(externalID string, reason string, adminUserID uint) error
	SearchListings(query string) ([]SearchResult, error)
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(claim *ListingClaim) error {
	return r.db.Create(claim).Error
}

func (r *GormRepository) FindByExternalID(externalID string) (*ListingClaim, error) {
	var claim ListingClaim
	err := r.db.Where("external_id = ?", externalID).First(&claim).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

func (r *GormRepository) FindPending() ([]ListingClaim, error) {
	var claims []ListingClaim
	err := r.db.Where("status = ?", StatusPending).Order("submitted_at ASC, id ASC").Find(&claims).Error
	return claims, err
}

func (r *GormRepository) FindByBusiness(businessID uint) ([]ListingClaim, error) {
	var claims []ListingClaim
	err := r.db.Where("business_id = ?", businessID).Order("submitted_at DESC, id DESC").Find(&claims).Error
	return claims, err
}

// Approve approves a claim and links the listing to the claiming business in a
// single transaction. The AND business_id IS NULL guard on the listing UPDATE is
// the real conflict check: if the listing already has an owner, the UPDATE
// affects 0 rows and we refuse to reassign it.
func (r *GormRepository) Approve(externalID string, adminUserID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var claim ListingClaim
		if err := tx.Where("external_id = ? AND status = ?", externalID, StatusPending).First(&claim).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrClaimNotFound
			}
			return err
		}

		table, ok := listingTable[claim.ListingType]
		if !ok {
			return ErrListingNotOwnable
		}

		now := time.Now()
		if err := tx.Model(&claim).
			Updates(map[string]interface{}{
				"status":      StatusApproved,
				"reviewed_at": now,
				"reviewed_by": adminUserID,
			}).Error; err != nil {
			return err
		}

		res := tx.Exec(
			fmt.Sprintf("UPDATE %s SET business_id = ? WHERE external_id = ? AND business_id IS NULL", table),
			claim.BusinessID, claim.ListingExternalID,
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrListingAlreadyOwned
		}
		return nil
	})
}

func (r *GormRepository) Reject(externalID string, reason string, adminUserID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&ListingClaim{}).
			Where("external_id = ? AND status = ?", externalID, StatusPending).
			Updates(map[string]interface{}{
				"status":           StatusRejected,
				"rejection_reason": reason,
				"reviewed_at":      time.Now(),
				"reviewed_by":      adminUserID,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrClaimNotFound
		}
		return nil
	})
}

func (r *GormRepository) SearchListings(query string) ([]SearchResult, error) {
	results := []SearchResult{}
	q := "%" + query + "%"
	log.Printf("Searching listings with query: %s, q: %s", query, q)
	err := r.db.Raw(`
		SELECT listing_type, external_id, name FROM (
			SELECT 'destination' AS listing_type, external_id, name FROM destinations WHERE name ILIKE ? AND status != 'draft'
			UNION ALL
			SELECT 'hotel', external_id, name FROM hotels WHERE name ILIKE ?
			UNION ALL
			SELECT 'restaurant', external_id, name FROM restaurants WHERE name ILIKE ?
			UNION ALL
			SELECT 'souvenir', external_id, name FROM souvenirs WHERE name ILIKE ?
			UNION ALL
			SELECT 'rental', external_id, name FROM rentals WHERE name ILIKE ?
			UNION ALL
			SELECT 'guide', external_id, name FROM guides WHERE name ILIKE ?
			UNION ALL
			SELECT 'event', external_id, title AS name FROM events WHERE title ILIKE ? AND status != 'draft'
		) t LIMIT 20
	`, q, q, q, q, q, q, q).Scan(&results).Error
	if err != nil {
		log.Printf("SearchListings error: %v", err)
	} else {
		log.Printf("SearchListings found %d results", len(results))
	}
	return results, err
}
