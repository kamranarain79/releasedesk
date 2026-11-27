package build

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	domainaudit "github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/storage"
)

type UploadFile struct {
	Header *multipart.FileHeader
	Arch   values.Architecture
	Kind   values.ArtifactKind
}

func (s *service) UploadBuild(slug values.Slug, os values.OS, info build.BasicInfo, files []UploadFile, Metadata map[string]string, opts UploadOptions) (build.Build, error) {
	for _, item := range files {
		if !isValidFileType(os, item.Header.Filename) {
			return build.Build{}, errors.New("invalid file type for this platform")
		}
	}

	platformEntity, err := s.platformRepo.GetByAppSlugAndOS(slug, os)
	if err != nil {
		return build.Build{}, err
	}

	exists, err := s.buildRepo.ExistsByNumber(platformEntity.ID, info.Number)
	if err != nil {
		return build.Build{}, err
	}
	if exists {
		if !opts.Overwrite {
			return build.Build{}, errors.New("build number already exists")
		}
		if err = s.buildRepo.DeleteByPlatformAndNumber(platformEntity.ID, info.Number); err != nil {
			return build.Build{}, err
		}
	}

	info.PlatformID = platformEntity.ID
	info.CreatedBy = opts.Actor
	info.Status = build.StatusUnpublished
	buildEntity := build.NewBuild(info)
	buildEntity.Metadata = Metadata

	for _, item := range files {
		file, err := item.Header.Open()
		if err != nil {
			return build.Build{}, errors.New("failed to open file")
		}

		md5Sum, sha256Sum, sha512Sum, err := storage.CalculateChecksums(file)
		_ = file.Close()
		if err != nil {
			return build.Build{}, err
		}
		if opts.ExpectedSHA256 != "" && len(files) == 1 && !strings.EqualFold(opts.ExpectedSHA256, sha256Sum) {
			return build.Build{}, errors.New("sha256 does not match uploaded file")
		}

		file, err = item.Header.Open()
		if err != nil {
			return build.Build{}, errors.New("failed to open file")
		}
		filePath := storage.ConvertChecksumToPath(sha256Sum)
		err = storage.SaveFile(filePath, file)
		_ = file.Close()
		if err != nil {
			return build.Build{}, err
		}

		kind := item.Kind
		if kind == "" {
			kind = kindFromFilename(os, item.Header.Filename)
		}

		artifact := build.NewArtifact(item.Header.Filename, item.Header.Size, item.Arch, kind, md5Sum, sha256Sum, sha512Sum)
		buildEntity.Artifacts = append(buildEntity.Artifacts, artifact)
	}

	err = s.buildRepo.Save(&buildEntity)
	if err != nil {
		return build.Build{}, err
	}
	s.audits.Record(platformEntity.AppID, opts.Actor, "uploaded", domainaudit.BuildSubject(string(os), info.Number), "Uploaded "+string(os)+" build #"+info.Number)

	return buildEntity, err
}

func (s *service) PublishBuild(platformID int, number int, actor string) error {
	existing, err := s.buildRepo.GetByPlatformIDAndNumber(platformID, number)
	if err != nil {
		return err
	}
	if existing.Status == build.StatusPublished {
		return nil
	}
	if existing.Status != "" && existing.Status != build.StatusUnpublished {
		return errors.New("only an unpublished build can be published")
	}
	if err = s.buildRepo.SetStatus(existing.ID, build.StatusPublished); err != nil {
		return err
	}
	if platformEntity, err := s.platformRepo.GetByID(platformID); err == nil {
		s.audits.Record(platformEntity.AppID, actor, "published", domainaudit.BuildSubject(string(platformEntity.OS), existing.Number), "Published "+string(platformEntity.OS)+" build #"+existing.Number)
	}
	return nil
}

func (s *service) DeleteBuild(platformID int, number int, actor string) error {
	existing, err := s.buildRepo.GetByPlatformIDAndNumber(platformID, number)
	if err != nil {
		return err
	}
	platformEntity, _ := s.platformRepo.GetByID(platformID)
	checksums, err := s.buildRepo.DeleteByID(existing.ID)
	if err != nil {
		return err
	}
	if platformEntity.AppID != 0 {
		s.audits.Record(platformEntity.AppID, actor, "deleted", domainaudit.BuildSubject(string(platformEntity.OS), existing.Number), "Deleted "+string(platformEntity.OS)+" build #"+existing.Number)
	}
	for _, checksum := range checksums {
		remaining, err := s.buildRepo.CountArtifactsByChecksum(checksum)
		if err != nil || remaining > 0 || !storage.ValidChecksum(checksum) {
			continue
		}
		_ = os.Remove(storage.ConvertChecksumToPath(checksum))
	}
	return nil
}

func isValidFileType(os values.OS, fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, allowedExt := range values.AllowedExtensions(os) {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}

func kindFromFilename(os values.OS, filename string) values.ArtifactKind {
	lower := strings.ToLower(filename)
	switch {
	case strings.Contains(lower, "bootloader"):
		return values.KindBootloader
	case strings.Contains(lower, "partition"):
		return values.KindPartition
	case strings.HasSuffix(lower, ".hex") || strings.HasSuffix(lower, ".srec"):
		return values.KindHex
	case strings.HasSuffix(lower, ".raucb"):
		return values.KindRAUC
	case os.IsFirmware() && strings.Contains(lower, "app"):
		return values.KindApplication
	default:
		return values.KindFirmware
	}
}
