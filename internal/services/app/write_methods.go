package app

import (
	"errors"
	"os"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/storage"
)

func (s *service) CreateApp(userID int, info app.BasicInfo) (app.App, error) {
	info.SetupGuideCompleted = false
	info.Slug.Format()

	newApp := app.NewApp(info)

	err := newApp.IsValid()
	if err != nil {
		return app.App{}, err
	}

	err = s.appRepo.Save(&newApp)
	if err != nil {
		return app.App{}, err
	}
	if userID > 0 {
		_ = s.memberRepo.Save(&app.Member{AppID: newApp.ID, UserID: userID, Role: app.MemberAdmin})
	}
	return newApp, nil
}

func (s *service) DeleteApp(actorID int, slug values.Slug) error {
	ok, err := s.CanAdmin(actorID, slug)
	if err != nil || !ok {
		return errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return err
	}
	checksums, err := s.appRepo.Delete(entity)
	if err != nil {
		return err
	}
	for _, checksum := range checksums {
		if storage.ValidChecksum(checksum) {
			_ = os.Remove(storage.ConvertChecksumToPath(checksum))
		}
	}
	return nil
}
