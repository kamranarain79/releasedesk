package build

type BuildRepository interface {
	Save(build *Build) error
	FindByPlatformID(platformID int) ([]BasicInfo, error)
	FindByPlatformIDAndVersion(platformID int, version string) ([]BasicInfo, error)
	GetByPlatformIDAndNumber(platformID int, number int) (Build, error)
	GetByID(id int) (Build, error)
	GetArtifactByChecksum(checksum string) (Artifact, error)
}
