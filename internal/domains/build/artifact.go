package build

import (
	"path/filepath"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/fields"
	"github.com/brewbits-co/releasedesk/pkg/storage"
)

func NewArtifact(filename string, size int64, arch values.Architecture, kind values.ArtifactKind, md5 string, sha256 string, sha512 string) Artifact {
	if kind == "" {
		kind = values.KindFirmware
	}
	filePath := storage.ConvertChecksumToPath(sha256)

	return Artifact{
		Storable: fields.Storable{
			Filename: filename,
			MimeType: mimeFromFilename(filename),
			Size:     fields.FileSize(size),
			Path:     filePath,
		},
		Hashable: fields.Hashable{
			Md5:    md5,
			Sha256: sha256,
			Sha512: sha512,
		},
		ID:           0,
		BuildID:      0,
		Architecture: arch,
		Kind:         kind,
	}
}

func mimeFromFilename(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".bin", ".hex", ".srec", ".img", ".wic":
		return "application/octet-stream"
	case ".elf":
		return "application/x-elf"
	case ".raucb":
		return "application/x-rauc-bundle"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".apk":
		return "application/vnd.android.package-archive"
	case ".ipa":
		return "application/octet-stream"
	default:
		return "application/octet-stream"
	}
}

type Artifact struct {
	fields.Storable `xorm:"extends"`
	fields.Hashable `xorm:"extends"`
	ID              int                 `xorm:"pk autoincr"`
	BuildID         int                 `xorm:"not null unique(artifact_uidx)"`
	Architecture    values.Architecture `xorm:"not null unique(artifact_uidx)"`
	Kind            values.ArtifactKind `xorm:"not null default 'firmware' unique(artifact_uidx)"`
}
