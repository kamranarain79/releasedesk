package build

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/storage"
)

type UploadFile struct {
	Header *multipart.FileHeader
	Arch   values.Architecture
	Kind   values.ArtifactKind
}

func (s *service) UploadBuild(slug values.Slug, os values.OS, info build.BasicInfo, files []UploadFile, Metadata map[string]string) (build.Build, error) {
	for _, item := range files {
		if !isValidFileType(os, item.Header.Filename) {
			return build.Build{}, errors.New("invalid file type for this platform")
		}
	}

	platformEntity, err := s.platformRepo.GetByAppSlugAndOS(slug, os)
	if err != nil {
		return build.Build{}, err
	}

	info.PlatformID = platformEntity.ID
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

	return buildEntity, err
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
