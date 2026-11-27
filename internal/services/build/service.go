package build

import (
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"github.com/brewbits-co/releasedesk/internal/values"
)

// Service defines the interface for handling build-related use cases.
type Service interface {
	// UploadBuild uploads a new build for the specified platform, using the provided build information and files.
	UploadBuild(slug values.Slug, os values.OS, info build.BasicInfo, files []UploadFile, Metadata map[string]string) (build.Build, error)
	// GetPlatformBuilds retrieves the list of build from a specific Platform.
	GetPlatformBuilds(platformID int) ([]build.BasicInfo, error)
	// GetBuildDetails return the full information of a Build including its Artifacts and Metadata.
	GetBuildDetails(platformID int, number int) (build.Build, error)
	GetBuildByID(id int) (build.Build, error)
	GetArtifactByChecksum(checksum string) (build.Artifact, error)
	FindByPlatformAndVersion(platformID int, version string) ([]build.BasicInfo, error)
}

// NewBuildService initializes a new instance of the build Service using the provided dependencies.
func NewBuildService(buildRepo build.BuildRepository, platformRepo platform.PlatformRepository) Service {
	return &service{
		buildRepo:    buildRepo,
		platformRepo: platformRepo,
	}
}

// service implements the build.Service
type service struct {
	buildRepo    build.BuildRepository
	platformRepo platform.PlatformRepository
}
