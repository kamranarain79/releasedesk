package release

import (
	"time"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	domainaudit "github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/values"
)

// Service defines the interface for handling release-related use cases.
type Service interface {
	CreateRelease(slug values.Slug, info release.BasicInfo) (release.Release, error)
	PublishRelease(id int, actor string) error
	DeleteRelease(id int, actor string) error
	ListReleasesByChannel(appID int, channelID int) ([]release.BasicInfo, error)
	ListRecentReleases(appID int, limit int) ([]release.BasicInfo, error)
	GetReleaseSummary(appID int, version string) (release.Release, error)
	GetReleaseChannels(appID int) ([]release.Channel, error)
	SaveReleaseNotes(releaseID int, text string, changelogs []release.Changelog, actor string) (release.ReleaseNotes, error)
	GetReleaseNotes(releaseID int) (release.ReleaseNotes, error)
	GetReleaseByID(releaseID int) (release.Release, error)
	UpdateReleaseBasicInfo(info release.BasicInfo) (release.Release, error)
	ReplaceLinks(releaseID int, links []release.LinkedBuilds) error
	SaveTestPlan(releaseID int, overview string, methods []release.TestMethod, actor string) (release.TestPlan, error)
	GetTestPlan(releaseID int) (release.TestPlan, error)
	SaveAnnouncement(item release.Announcement) (release.Announcement, error)
	ListAnnouncements(appID int) ([]release.Announcement, error)
}

// NewReleaseService initializes a new instance of the release Service using the provided dependencies.
func NewReleaseService(releaseRepo release.ReleaseRepository, releaseNotesRepo release.ReleaseNotesRepository, testMethodRepo release.TestMethodRepository, appRepo app.AppRepository, announcementRepo release.AnnouncementRepository, audits auditsvc.Service) Service {
	return &service{
		releaseRepo:      releaseRepo,
		releaseNotesRepo: releaseNotesRepo,
		testMethodRepo:   testMethodRepo,
		appRepo:          appRepo,
		announcementRepo: announcementRepo,
		audits:           audits,
	}
}

type service struct {
	releaseRepo      release.ReleaseRepository
	releaseNotesRepo release.ReleaseNotesRepository
	testMethodRepo   release.TestMethodRepository
	appRepo          app.AppRepository
	announcementRepo release.AnnouncementRepository
	audits           auditsvc.Service
}

// SaveReleaseNotes saves release notes and associated changelogs for a release
func (s *service) SaveReleaseNotes(releaseID int, text string, changelogs []release.Changelog, actor string) (release.ReleaseNotes, error) {
	releaseNotes := release.NewReleaseNotes(releaseID, text)
	releaseNotes.Changelogs = changelogs

	if err := s.releaseNotesRepo.Save(&releaseNotes); err != nil {
		return release.ReleaseNotes{}, err
	}
	if existing, err := s.releaseRepo.GetByID(releaseID); err == nil {
		s.audits.Record(existing.AppID, actor, "edited", domainaudit.ReleaseSubject(existing.Version), "Edited release notes for "+existing.Version)
	}

	return releaseNotes, nil
}

// GetReleaseNotes retrieves release notes and associated changelogs for a release
func (s *service) GetReleaseNotes(releaseID int) (release.ReleaseNotes, error) {
	releaseNotes, err := s.releaseNotesRepo.GetByReleaseID(releaseID)
	if err != nil {
		return release.ReleaseNotes{}, err
	}

	return releaseNotes, nil
}

func (s *service) SaveTestPlan(releaseID int, overview string, methods []release.TestMethod, actor string) (release.TestPlan, error) {
	if err := s.testMethodRepo.SavePlan(releaseID, overview, methods); err != nil {
		return release.TestPlan{}, err
	}
	if existing, err := s.releaseRepo.GetByID(releaseID); err == nil {
		s.audits.Record(existing.AppID, actor, "edited", domainaudit.ReleaseSubject(existing.Version), "Edited QA tests for "+existing.Version)
	}
	return s.testMethodRepo.GetPlan(releaseID)
}

func (s *service) GetTestPlan(releaseID int) (release.TestPlan, error) {
	return s.testMethodRepo.GetPlan(releaseID)
}

func (s *service) SaveAnnouncement(item release.Announcement) (release.Announcement, error) {
	item.CreatedAt = item.CreatedAt.UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if err := s.announcementRepo.Save(&item); err != nil {
		return release.Announcement{}, err
	}
	return item, nil
}

func (s *service) ListAnnouncements(appID int) ([]release.Announcement, error) {
	return s.announcementRepo.ListByApp(appID)
}
