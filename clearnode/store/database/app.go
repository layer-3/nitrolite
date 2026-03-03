package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"gorm.io/gorm"
)

// AppV1 represents an application registry entry in the database.
type AppV1 struct {
	ID                          string `gorm:"primaryKey"`
	OwnerWallet                 string `gorm:"column:owner_wallet;not null"`
	Metadata                    string `gorm:"column:metadata;type:text;not null"`
	Version                     uint64 `gorm:"column:version;default:1"`
	CreationApprovalNotRequired bool   `gorm:"column:creation_approval_not_required"`
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

func (AppV1) TableName() string {
	return "apps_v1"
}

// CreateApp registers a new application. Returns an error if the app ID already exists.
func (s *DBStore) CreateApp(entry app.AppV1) error {
	dbApp := AppV1{
		ID:                          strings.ToLower(entry.ID),
		OwnerWallet:                 strings.ToLower(entry.OwnerWallet),
		Metadata:                    entry.Metadata,
		Version:                     entry.Version,
		CreationApprovalNotRequired: entry.CreationApprovalNotRequired,
	}

	if err := s.db.Create(&dbApp).Error; err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	return nil
}

// GetApp retrieves a single application by ID. Returns nil if not found.
func (s *DBStore) GetApp(appID string) (*app.AppInfoV1, error) {
	var dbApp AppV1
	err := s.db.Where("id = ?", strings.ToLower(appID)).First(&dbApp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	result := databaseAppToCore(&dbApp)
	return &result, nil
}

// GetApps retrieves applications with optional filtering by app ID, owner wallet, and pagination.
func (s *DBStore) GetApps(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
	query := s.db.Model(&AppV1{})

	if appID != nil && *appID != "" {
		query = query.Where("id = ?", strings.ToLower(*appID))
	}

	if ownerWallet != nil && *ownerWallet != "" {
		query = query.Where("owner_wallet = ?", strings.ToLower(*ownerWallet))
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to count apps: %w", err)
	}

	offset, limit := pagination.GetOffsetAndLimit(DefaultLimit, MaxLimit)

	query = query.Order("created_at DESC").Offset(int(offset)).Limit(int(limit))

	var dbApps []AppV1
	if err := query.Find(&dbApps).Error; err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get apps: %w", err)
	}

	apps := make([]app.AppInfoV1, len(dbApps))
	for i, dbApp := range dbApps {
		apps[i] = databaseAppToCore(&dbApp)
	}

	metadata := calculatePaginationMetadata(totalCount, offset, limit)

	return apps, metadata, nil
}

func databaseAppToCore(dbApp *AppV1) app.AppInfoV1 {
	return app.AppInfoV1{
		App: app.AppV1{
			ID:                          dbApp.ID,
			OwnerWallet:                 dbApp.OwnerWallet,
			Metadata:                    dbApp.Metadata,
			Version:                     dbApp.Version,
			CreationApprovalNotRequired: dbApp.CreationApprovalNotRequired,
		},
		CreatedAt: dbApp.CreatedAt,
		UpdatedAt: dbApp.UpdatedAt,
	}
}
