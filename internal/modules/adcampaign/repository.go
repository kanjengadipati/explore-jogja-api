package adcampaign

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]AdCampaign, error)
	FindByID(externalID string) (*AdCampaign, error)
	FindActiveCandidates(placement, category string) ([]AdCampaign, error)
	FindAllByBusinessExternalID(businessExternalID string) ([]AdCampaign, error)
	BusinessExists(externalID string) (bool, error)
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

func (r *GormRepository) BusinessExists(externalID string) (bool, error) {
	var count int64
	err := r.db.Table("businesses").
		Where("external_id = ?", externalID).
		Count(&count).Error
	return count > 0, err
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
