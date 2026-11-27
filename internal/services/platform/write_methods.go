package platform

import (
	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"github.com/brewbits-co/releasedesk/internal/values"
)

func (s *service) AddPlatformToApp(slug values.Slug, info platform.BasicInfo) (platform.Platform, error) {
	appEntity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return platform.Platform{}, app.ErrAppNotFound
	}

	info.AppID = appEntity.ID
	addedPlatform := platform.NewPlatform(info)

	err = s.platformRepo.Save(&addedPlatform)
	if err != nil {
		return platform.Platform{}, err
	}

	return addedPlatform, nil
}

func (s *service) SaveConfig(slug values.Slug, os values.OS, key, value string) error {
	entity, err := s.platformRepo.GetByAppSlugAndOS(slug, os)
	if err != nil {
		return err
	}
	return s.configRepo.Save(platform.Config{PlatformID: entity.ID, Key: key, Value: value})
}

func (s *service) ListConfig(slug values.Slug, os values.OS) ([]platform.Config, error) {
	entity, err := s.platformRepo.GetByAppSlugAndOS(slug, os)
	if err != nil {
		return nil, err
	}
	return s.configRepo.List(entity.ID)
}

func (s *service) FlashNotes(platformID int) string {
	items, err := s.configRepo.List(platformID)
	if err != nil {
		return ""
	}
	for _, item := range items {
		if item.Key == "flash_notes" {
			return item.Value
		}
	}
	return ""
}
