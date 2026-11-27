package app

import (
	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

// Service defines the interface for handling application-related use cases.
type Service interface {
	// CreateApp creates a new application and adds the creator as app admin.
	// It validates the input data against defined business rules and returns
	// the created application or an error if the validation or creation fails.
	CreateApp(userID int, info app.BasicInfo) (app.App, error)
	DeleteApp(actorID int, slug values.Slug) error
	// GetUserAccessibleApps retrieves the list of applications that a given user
	// has access to, based on the user's ID.
	GetUserAccessibleApps(userID int) ([]app.App, error)
	// ApplyAppSetupGuide applies the setup guide for the specified application,
	// configuring it with the chosen settings.
	ApplyAppSetupGuide(slug values.Slug, format values.VersionFormat, channels app.SetupChannelsOption, customChannels []string) error
	// GetAppOverview returns a summary of an application.
	GetAppOverview(slug values.Slug) (app.Overview, error)
	// GetCurrentAppData retrieves application information shared across views based on the provided slug.
	GetCurrentAppData(slug values.Slug) (session.CurrentAppData, error)
	// GetPublicApps returns apps that are visible on the download portal.
	GetPublicApps() ([]app.App, error)
	GetBySlug(slug values.Slug) (app.App, error)
	GetByID(id int) (app.App, error)
	CanView(userID int, slug values.Slug) (bool, error)
	CanAdmin(userID int, slug values.Slug) (bool, error)
	CanEditContent(userID int, slug values.Slug) (bool, error)
	UpdateSettings(userID int, slug values.Slug, description string, private bool, teamsWebhook string) (app.App, error)
	ListMembers(userID int, slug values.Slug) ([]app.Member, error)
	AddMember(actorID int, slug values.Slug, userID int, role app.MemberRole) error
	RemoveMember(actorID int, slug values.Slug, userID int) error
	MintToken(actorID int, slug values.Slug, name, scope string) (raw string, token app.CIToken, err error)
	ListTokens(actorID int, slug values.Slug) ([]app.CIToken, error)
	RevokeToken(actorID int, slug values.Slug, tokenID int) error
	LookupToken(raw string) (app.CIToken, error)
	IsInstanceAdmin(userID int) bool
	ListGroups(userID int, slug values.Slug) ([]app.DistributionGroup, error)
	SaveGroup(userID int, slug values.Slug, name string, restrict bool) error
}

// NewAppService initializes a new instance of the application Service using the provided dependencies.
func NewAppService(appRepo app.AppRepository, platformRepo platform.PlatformRepository, memberRepo app.MemberRepository, tokenRepo app.CITokenRepository, userRepo user.UserRepository, groupRepo app.GroupRepository) Service {
	return &service{
		appRepo:      appRepo,
		platformRepo: platformRepo,
		memberRepo:   memberRepo,
		tokenRepo:    tokenRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
	}
}

// service implements the app.Service
type service struct {
	appRepo      app.AppRepository
	platformRepo platform.PlatformRepository
	memberRepo   app.MemberRepository
	tokenRepo    app.CITokenRepository
	userRepo     user.UserRepository
	groupRepo    app.GroupRepository
}
