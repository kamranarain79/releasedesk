package release

import (
	"errors"

	domainaudit "github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
)

func (s *service) CreateRelease(slug values.Slug, info release.BasicInfo) (release.Release, error) {
	appEntity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return release.Release{}, err
	}

	info.AppID = appEntity.ID
	info.Status = release.Unpublished
	info.BuildSelection = release.Last

	newRelease := release.NewRelease(info)

	err = s.releaseRepo.Save(&newRelease)
	if err != nil {
		return release.Release{}, err
	}
	s.audits.Record(newRelease.AppID, info.CreatedBy, "created", domainaudit.ReleaseSubject(newRelease.Version), "Created release "+newRelease.Version)

	return newRelease, nil
}

func (s *service) PublishRelease(id int, actor string) error {
	existing, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing.Status == release.Published {
		return nil
	}
	if existing.Status != release.Unpublished {
		return errors.New("only an unpublished release can be published")
	}
	if err = s.releaseRepo.SetStatus(id, release.Published); err != nil {
		return err
	}
	s.audits.Record(existing.AppID, actor, "published", domainaudit.ReleaseSubject(existing.Version), "Published release "+existing.Version)
	return nil
}

func (s *service) DeleteRelease(id int, actor string) error {
	existing, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return err
	}
	if err = s.releaseRepo.Delete(id); err != nil {
		return err
	}
	s.audits.Record(existing.AppID, actor, "deleted", domainaudit.ReleaseSubject(existing.Version), "Deleted release "+existing.Version)
	return nil
}
