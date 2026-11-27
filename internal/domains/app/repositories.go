package app

import "github.com/brewbits-co/releasedesk/internal/values"

type AppRepository interface {
	Save(app *App) error
	Find() ([]App, error)
	GetBySlug(slug values.Slug) (App, error)
	GetByID(id int) (App, error)
	Update(app App) error
	Delete(app App) (checksums []string, err error)
	GetPlatformAvailability(app *App) error
	SaveSetupGuide(guide SetupGuide) error
	UpdateSettings(id int, description string, private bool, teamsWebhook string) error
}
