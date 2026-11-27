package branding

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brewbits-co/releasedesk/pkg/storage"
)

const maxImage = 2 << 20

var exts = []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico"}

func dir() string {
	return filepath.Join(filepath.Dir(storage.Root), "branding")
}

// SaveImage writes a company logo or favicon, replacing any previous file of that kind.
func SaveImage(kind string, r io.Reader, claimedType string) error {
	if kind != "logo" && kind != "favicon" {
		return errors.New("unknown image")
	}
	return writeFile(filepath.Join(dir(), kind), r, claimedType)
}

// SaveAvatar writes the signed-in user's photo.
func SaveAvatar(userID int, r io.Reader, claimedType string) error {
	if userID <= 0 {
		return errors.New("unknown user")
	}
	return writeFile(filepath.Join(dir(), "avatars", strconv.Itoa(userID)), r, claimedType)
}

func writeFile(pathNoExt string, r io.Reader, claimedType string) error {
	buf, err := io.ReadAll(io.LimitReader(r, maxImage+1))
	if err != nil {
		return err
	}
	if len(buf) == 0 || len(buf) > maxImage {
		return errors.New("image must be between 1 byte and 2MB")
	}
	kind := http.DetectContentType(buf)
	if claimedType != "" && !strings.HasPrefix(claimedType, "image/") && claimedType != "application/octet-stream" {
		return errors.New("file is not an image")
	}
	ext := extFor(kind)
	if ext == "" {
		return errors.New("use PNG, JPEG, GIF, WebP, or ICO")
	}
	if err = os.MkdirAll(filepath.Dir(pathNoExt), 0o750); err != nil {
		return err
	}
	clearExts(pathNoExt)
	return os.WriteFile(pathNoExt+ext, buf, 0o640)
}

func extFor(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	default:
		return ""
	}
}

func clearExts(pathNoExt string) {
	for _, ext := range exts {
		_ = os.Remove(pathNoExt + ext)
	}
}

// Remove deletes every stored file for a logo, favicon, or avatar path without an extension.
func Remove(pathNoExt string) {
	clearExts(pathNoExt)
}

// RemoveBrand deletes the company logo or favicon.
func RemoveBrand(kind string) {
	if kind == "logo" || kind == "favicon" {
		clearExts(filepath.Join(dir(), kind))
	}
}

// RemoveAvatar deletes a user's photo.
func RemoveAvatar(userID int) {
	if userID > 0 {
		clearExts(filepath.Join(dir(), "avatars", strconv.Itoa(userID)))
	}
}

// Open returns the first stored file for a path without an extension.
func Open(pathNoExt string) (*os.File, string, error) {
	for _, ext := range exts {
		file, err := os.Open(pathNoExt + ext)
		if err == nil {
			return file, contentType(ext), nil
		}
	}
	return nil, "", os.ErrNotExist
}

// OpenBrand opens the company logo or favicon.
func OpenBrand(kind string) (*os.File, string, error) {
	if kind != "logo" && kind != "favicon" {
		return nil, "", os.ErrNotExist
	}
	return Open(filepath.Join(dir(), kind))
}

// OpenAvatar opens a user's photo.
func OpenAvatar(userID int) (*os.File, string, error) {
	return Open(filepath.Join(dir(), "avatars", strconv.Itoa(userID)))
}

func contentType(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}
