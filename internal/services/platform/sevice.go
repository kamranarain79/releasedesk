package platform

import (
	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"github.com/brewbits-co/releasedesk/internal/values"
)

// Service defines the interface for handling platform-related use cases.
type Service interface {
	// AddPlatformToApp adds a new platform for a specific application using the provided basic information.
	// It validates the input data against defined business rules and returns
	// the created platform or an error if the validation or creation fails.
	AddPlatformToApp(slug values.Slug, info platform.BasicInfo) (platform.Platform, error)
	SaveConfig(slug values.Slug, os values.OS, key, value string) error
	ListConfig(slug values.Slug, os values.OS) ([]platform.Config, error)
	FlashNotes(platformID int) string
}

// NewPlatformService initializes a new instance of the platform Service using the provided dependencies.
func NewPlatformService(platformRepo platform.PlatformRepository, appRepo app.AppRepository, configRepo platform.ConfigRepository) Service {
	return &service{
		platformRepo: platformRepo,
		appRepo:      appRepo,
		configRepo:   configRepo,
	}
}

// service implements the platform.Service
type service struct {
	platformRepo platform.PlatformRepository
	appRepo      app.AppRepository
	configRepo   platform.ConfigRepository
}
