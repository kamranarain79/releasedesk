package build

type BuildRepository interface {
	Save(build *Build) error
	DeleteByPlatformAndNumber(platformID int, number string) error
	ExistsByNumber(platformID int, number string) (bool, error)
	FindByPlatformID(platformID int) ([]BasicInfo, error)
	FindByPlatformIDAndVersion(platformID int, version string) ([]BasicInfo, error)
	GetByPlatformIDAndNumber(platformID int, number int) (Build, error)
	GetByID(id int) (Build, error)
	GetArtifactByChecksum(checksum string) (Artifact, error)
	SetStatus(id int, status string) error
	DeleteByID(id int) (checksums []string, err error)
	CountArtifactsByChecksum(checksum string) (int64, error)
}
