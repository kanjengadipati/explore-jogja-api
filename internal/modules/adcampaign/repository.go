package adcampaign

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"pleco-api/internal/utils"
)

type Repository interface {
	FindAll() ([]AdCampaign, error)
	FindByID(externalID string) (*AdCampaign, error)
	FindActiveCandidates(placement, category string) ([]AdCampaign, error)
	FindActiveEcosystemCandidates(destinationID string) ([]AdCampaign, error)
	FindEcosystemListing(listingType, externalID string) (*EcosystemListing, error)
	FindDestinationCoords(externalID string) (lat, lng float64, found bool)
	ListingBelongsToBusiness(listingType, listingExternalID, businessExternalID string) (bool, error)
	FindAllByBusinessExternalID(businessExternalID string) ([]AdCampaign, error)
	BusinessExists(externalID string) (bool, error)
	OwnerEmailsForBusiness(businessExternalID string) ([]string, error)
	UserEmailByID(userID uint) (string, error)
	Create(campaign *AdCampaign) error
	Update(campaign *AdCampaign) error
	Delete(externalID string) error
	IncrementImpression(externalID string) error
	IncrementClick(externalID string) error
	FindAllHouseAds() ([]HouseAd, error)
	FindEnabledHouseAdByPlacement(placement string) (*HouseAd, error)
	FindHouseAdByID(externalID string) (*HouseAd, error)
	CreateHouseAd(houseAd *HouseAd) error
	UpdateHouseAd(houseAd *HouseAd) error
	DeleteHouseAd(externalID string) error
	FindPlacementPrice(placement string) (*AdPlacementPrice, error)
	FindAllPlacementPrices() ([]AdPlacementPrice, error)
	UpsertPlacementPrice(price *AdPlacementPrice) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAllByBusinessExternalID(businessExternalID string) ([]AdCampaign, error) {
	var campaigns []AdCampaign
	err := r.db.Table("ad_campaigns").
		Select("ad_campaigns.*, businesses.name AS business_name").
		Joins("LEFT JOIN businesses ON businesses.external_id = ad_campaigns.business_external_id").
		Where("ad_campaigns.business_external_id = ?", businessExternalID).
		Order("ad_campaigns.id DESC").
		Scan(&campaigns).Error
	return campaigns, err
}

func (r *GormRepository) FindAll() ([]AdCampaign, error) {
	var campaigns []AdCampaign
	err := r.db.Table("ad_campaigns").
		Select("ad_campaigns.*, businesses.name AS business_name").
		Joins("LEFT JOIN businesses ON businesses.external_id = ad_campaigns.business_external_id").
		Order("ad_campaigns.id DESC").
		Scan(&campaigns).Error
	return campaigns, err
}

func (r *GormRepository) FindByID(externalID string) (*AdCampaign, error) {
	var campaign AdCampaign
	err := r.db.Table("ad_campaigns").
		Select("ad_campaigns.*, businesses.name AS business_name").
		Joins("LEFT JOIN businesses ON businesses.external_id = ad_campaigns.business_external_id").
		Where("ad_campaigns.external_id = ?", externalID).
		Limit(1).
		Scan(&campaign).Error
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *GormRepository) FindActiveCandidates(placement, category string) ([]AdCampaign, error) {
	now := time.Now()
	zero := time.Time{}
	q := r.db.Table("ad_campaigns").
		Select("ad_campaigns.*, businesses.name AS business_name").
		Joins("LEFT JOIN businesses ON businesses.external_id = ad_campaigns.business_external_id").
		Where("ad_campaigns.placement = ?", placement).
		Where("ad_campaigns.is_active = ?", true).
		Where("ad_campaigns.payment_status = ?", "paid").
		Where("(ad_campaigns.start_at IS NULL OR ad_campaigns.start_at = ? OR ad_campaigns.start_at <= ?)", zero, now).
		Where("(ad_campaigns.end_at IS NULL OR ad_campaigns.end_at = ? OR ad_campaigns.end_at >= ?)", zero, now)

	if category != "" {
		q = q.Where("(ad_campaigns.category = ? OR ad_campaigns.category = '')", category)
	}

	var campaigns []AdCampaign
	err := q.Scan(&campaigns).Error
	return campaigns, err
}

// FindActiveEcosystemCandidates returns live paid campaigns on any ecosystem_*
// placement that target the given destination (or all destinations when the
// campaign's target_dest_ids is empty). Ordered by sort_order (lower first) so
// the rail shows higher-paying listings on top.
func (r *GormRepository) FindActiveEcosystemCandidates(destinationID string) ([]AdCampaign, error) {
	now := time.Now()
	zero := time.Time{}
	q := r.db.Table("ad_campaigns").
		Select("ad_campaigns.*, businesses.name AS business_name").
		Joins("LEFT JOIN businesses ON businesses.external_id = ad_campaigns.business_external_id").
		Where("ad_campaigns.placement IN ?", EcosystemPlacements).
		Where("ad_campaigns.is_active = ?", true).
		Where("ad_campaigns.payment_status = ?", "paid").
		Where("(ad_campaigns.start_at IS NULL OR ad_campaigns.start_at = ? OR ad_campaigns.start_at <= ?)", zero, now).
		Where("(ad_campaigns.end_at IS NULL OR ad_campaigns.end_at = ? OR ad_campaigns.end_at >= ?)", zero, now)

	if destinationID != "" {
		q = q.Where("ad_campaigns.target_dest_ids = '[]'::jsonb OR ad_campaigns.target_dest_ids @> ?", `["`+destinationID+`"]`)
	}

	var campaigns []AdCampaign
	err := q.Order("ad_campaigns.sort_order ASC, ad_campaigns.id ASC").Scan(&campaigns).Error
	return campaigns, err
}

// EcosystemListing is the denormalized listing card data used to enrich an
// ecosystem campaign at serve time. It maps the heterogeneous listing tables
// (hotels / restaurants / souvenirs / rentals / guides) onto one shape.
type EcosystemListing struct {
	Name        string
	Description string
	Address     string
	Image       string
	Rating      float64
	Price       string
	Phone       string
	Website     string
	Latitude    float64
	Longitude   float64
}

// FindEcosystemListing loads the card fields for a listing promoted by an
// ecosystem campaign. ListingType is one of hotel | restaurant | souvenir |
// rental | guide (cafe maps to the restaurants table, transport to rentals).
func (r *GormRepository) FindEcosystemListing(listingType, externalID string) (*EcosystemListing, error) {
	type raw struct {
		Name        string
		Description string
		Address     string
		Image       string
		Images      utils.JSONArr
		Rating      float64
		Price       string
		Phone       string
		Website     string
		Latitude    float64
		Longitude   float64
	}

	var row raw
	var err error

	switch listingType {
	case "hotel":
		err = r.db.Table("hotels").
			Select("name, description, address, images, rating, price_per_night AS price, phone, website, latitude, longitude").
			Where("external_id = ?", externalID).Limit(1).Scan(&row).Error
	case "restaurant", "cafe":
		err = r.db.Table("restaurants").
			Select("name, description, address, images, rating, price_range AS price, phone, latitude, longitude").
			Where("external_id = ?", externalID).Limit(1).Scan(&row).Error
	case "souvenir":
		err = r.db.Table("souvenirs").
			Select("name, description, address, images, rating, price_range AS price, phone, latitude, longitude").
			Where("external_id = ?", externalID).Limit(1).Scan(&row).Error
	case "rental", "transport":
		err = r.db.Table("rentals").
			Select("name, description, address, images, rating, price_per_day AS price, phone, latitude, longitude").
			Where("external_id = ?", externalID).Limit(1).Scan(&row).Error
	case "guide":
		err = r.db.Table("guides").
			Select("name, bio AS description, avatar AS image, rating, price_per_day AS price, phone").
			Where("external_id = ?", externalID).Limit(1).Scan(&row).Error
	default:
		return nil, fmt.Errorf("unsupported listing_type: %s", listingType)
	}
	if err != nil {
		return nil, err
	}

	image := row.Image
	if image == "" && len(row.Images) > 0 {
		if s, ok := row.Images[0].(string); ok {
			image = s
		}
	}

	return &EcosystemListing{
		Name:        row.Name,
		Description: row.Description,
		Address:     row.Address,
		Image:       image,
		Rating:      row.Rating,
		Price:       row.Price,
		Phone:       row.Phone,
		Website:     row.Website,
		Latitude:    row.Latitude,
		Longitude:   row.Longitude,
	}, nil
}

// FindDestinationCoords returns the lat/lng of a destination used to compute
// the card's distance field. `found` is false when the destination is unknown.
func (r *GormRepository) FindDestinationCoords(externalID string) (lat, lng float64, found bool) {
	var row struct {
		Latitude  float64
		Longitude float64
	}
	err := r.db.Table("destinations").
		Select("latitude, longitude").
		Where("external_id = ?", externalID).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return 0, 0, false
	}
	return row.Latitude, row.Longitude, true
}

// listingTable maps an ecosystem listing_type to its DB table name.
func listingTable(listingType string) string {
	switch listingType {
	case "hotel":
		return "hotels"
	case "restaurant", "cafe":
		return "restaurants"
	case "souvenir":
		return "souvenirs"
	case "rental", "transport":
		return "rentals"
	case "guide":
		return "guides"
	default:
		return ""
	}
}

// ListingBelongsToBusiness reports whether the listing exists AND is owned by
// the given business (business_id FK set by the listing-claim approval flow).
func (r *GormRepository) ListingBelongsToBusiness(listingType, listingExternalID, businessExternalID string) (bool, error) {
	table := listingTable(listingType)
	if table == "" {
		return false, fmt.Errorf("unsupported listing_type: %s", listingType)
	}
	var count int64
	err := r.db.Table(table).
		Joins("JOIN businesses b ON b.id = "+table+".business_id").
		Where(table+".external_id = ?", listingExternalID).
		Where("b.external_id = ?", businessExternalID).
		Where(table+".business_id IS NOT NULL").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormRepository) BusinessExists(externalID string) (bool, error) {
	var count int64
	err := r.db.Table("businesses").
		Where("external_id = ?", externalID).
		Count(&count).Error
	return count > 0, err
}

// OwnerEmailsForBusiness resolves the recipient email addresses for approval /
// rejection notifications of a campaign owned by a business. It collects the
// unique emails of the user accounts linked via business_owners, falling back
// to the business contact email when no owner user has an email on file.
func (r *GormRepository) OwnerEmailsForBusiness(businessExternalID string) ([]string, error) {
	var emails []string
	err := r.db.Raw(`
		SELECT DISTINCT u.email
		FROM business_owners bo
		JOIN users u ON u.id = bo.user_id
		JOIN businesses b ON b.id = bo.business_id
		WHERE b.external_id = ?
		  AND u.email <> ''`, businessExternalID).Scan(&emails).Error
	if err != nil {
		return nil, err
	}
	if len(emails) > 0 {
		return emails, nil
	}
	err = r.db.Table("businesses").
		Select("email").
		Where("external_id = ? AND email <> ''", businessExternalID).
		Scan(&emails).Error
	return emails, err
}

// UserEmailByID returns the email of a user account, used to record the acting
// admin's identity in approved_by / rejected_by.
func (r *GormRepository) UserEmailByID(userID uint) (string, error) {
	var email string
	err := r.db.Table("users").
		Where("id = ?", userID).
		Pluck("email", &email).Error
	return email, err
}

func (r *GormRepository) Create(campaign *AdCampaign) error {
	return r.db.Create(campaign).Error
}

func (r *GormRepository) Update(campaign *AdCampaign) error {
	return r.db.Save(campaign).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&AdCampaign{}).Error
}

func (r *GormRepository) IncrementImpression(externalID string) error {
	return r.db.Model(&AdCampaign{}).
		Where("external_id = ?", externalID).
		UpdateColumn("impressions", gorm.Expr("impressions + 1")).Error
}

func (r *GormRepository) IncrementClick(externalID string) error {
	return r.db.Model(&AdCampaign{}).
		Where("external_id = ?", externalID).
		UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
}

func (r *GormRepository) FindAllHouseAds() ([]HouseAd, error) {
	var houseAds []HouseAd
	err := r.db.Order("placement ASC").Find(&houseAds).Error
	return houseAds, err
}

func (r *GormRepository) FindEnabledHouseAdByPlacement(placement string) (*HouseAd, error) {
	var houseAd HouseAd
	err := r.db.Where("placement = ? AND is_enabled = ?", placement, true).First(&houseAd).Error
	if err != nil {
		return nil, err
	}
	return &houseAd, nil
}

func (r *GormRepository) FindHouseAdByID(externalID string) (*HouseAd, error) {
	var houseAd HouseAd
	err := r.db.Where("external_id = ?", externalID).First(&houseAd).Error
	if err != nil {
		return nil, err
	}
	return &houseAd, nil
}

func (r *GormRepository) CreateHouseAd(houseAd *HouseAd) error {
	return r.db.Create(houseAd).Error
}

func (r *GormRepository) UpdateHouseAd(houseAd *HouseAd) error {
	return r.db.Save(houseAd).Error
}

func (r *GormRepository) DeleteHouseAd(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&HouseAd{}).Error
}

func (r *GormRepository) FindPlacementPrice(placement string) (*AdPlacementPrice, error) {
	var price AdPlacementPrice
	err := r.db.Where("placement = ?", placement).First(&price).Error
	if err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *GormRepository) FindAllPlacementPrices() ([]AdPlacementPrice, error) {
	var prices []AdPlacementPrice
	err := r.db.Order("placement ASC").Find(&prices).Error
	return prices, err
}

func (r *GormRepository) UpsertPlacementPrice(price *AdPlacementPrice) error {
	return r.db.Save(price).Error
}

func WeightedPick(candidates []AdCampaign) *AdCampaign {
	if len(candidates) == 0 {
		return nil
	}
	total := 0
	for _, c := range candidates {
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}
	r := rand.Intn(total)
	for i := range candidates {
		w := candidates[i].Weight
		if w <= 0 {
			w = 1
		}
		if r < w {
			return &candidates[i]
		}
		r -= w
	}
	return &candidates[len(candidates)-1]
}
