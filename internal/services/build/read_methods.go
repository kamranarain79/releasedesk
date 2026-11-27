package build

import "github.com/brewbits-co/releasedesk/internal/domains/build"

func (s *service) GetPlatformBuilds(platformID int) ([]build.BasicInfo, error) {
	builds, err := s.buildRepo.FindByPlatformID(platformID)
	if err != nil {
		return nil, err
	}

	return builds, nil
}

func (s *service) GetBuildDetails(platformID int, number int) (build.Build, error) {
	buildEntity, err := s.buildRepo.GetByPlatformIDAndNumber(platformID, number)
	if err != nil {
		return build.Build{}, err
	}

	return buildEntity, err
}

func (s *service) GetBuildByID(id int) (build.Build, error) {
	return s.buildRepo.GetByID(id)
}

func (s *service) GetArtifactByChecksum(checksum string) (build.Artifact, error) {
	return s.buildRepo.GetArtifactByChecksum(checksum)
}

func (s *service) FindByPlatformAndVersion(platformID int, version string) ([]build.BasicInfo, error) {
	return s.buildRepo.FindByPlatformIDAndVersion(platformID, version)
}

func (s *service) IncrementDownload(sha256 string) error {
	return s.downloads.Increment(sha256)
}

func (s *service) SumDownloads(checksums []string) (int64, error) {
	return s.downloads.Sum(checksums)
}
