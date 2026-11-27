package build

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	domainbuild "github.com/brewbits-co/releasedesk/internal/domains/build"
	buildsrv "github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/fields"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/brewbits-co/releasedesk/pkg/storage"
	"github.com/brewbits-co/releasedesk/pkg/teams"
	"github.com/brewbits-co/releasedesk/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func maxUploadBytes() int64 {
	mb := 512
	if raw := os.Getenv("RELEASEDESK_MAX_UPLOAD_MB"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			mb = n
		}
	}
	return int64(mb) << 20
}

func (c *buildController) HandleBuildUpload(w http.ResponseWriter, r *http.Request) {
	limit := maxUploadBytes()
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseMultipartForm(limit); err != nil {
		http.Error(w, "upload too large or malformed", http.StatusBadRequest)
		return
	}

	slug := chi.URLParam(r, "slug")
	platform := chi.URLParam(r, "platform")

	var buildUploadRequest = struct {
		domainbuild.BasicInfo
		fields.Extendable
	}{}

	err := utils.NewDecoder().Decode(&buildUploadRequest, r.Form)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.PlainText(w, r, err.Error())
		return
	}

	var files []buildsrv.UploadFile
	appendFile := func(formName string, arch values.Architecture, kind values.ArtifactKind) {
		if r.MultipartForm == nil {
			return
		}
		if fhs := r.MultipartForm.File[formName]; len(fhs) > 0 {
			files = append(files, buildsrv.UploadFile{Header: fhs[0], Arch: arch, Kind: kind})
		}
	}

	defaultKind := values.KindApplication
	if values.OS(platform).IsFirmware() {
		defaultKind = values.KindFirmware
	}
	appendFile("File", values.NoArch, defaultKind)
	appendFile("File_firmware", values.NoArch, values.KindFirmware)
	appendFile("File_bootloader", values.NoArch, values.KindBootloader)
	appendFile("File_partition", values.NoArch, values.KindPartition)
	appendFile("File_application", values.NoArch, values.KindApplication)
	appendFile("File_hex", values.NoArch, values.KindHex)
	appendFile("File_rauc", values.NoArch, values.KindRAUC)
	appendFile("File_x86", values.X86, defaultKind)
	appendFile("File_x64", values.X64, defaultKind)
	appendFile("File_ARM64", values.ARM64, defaultKind)
	appendFile("File_ARM", values.ARM, defaultKind)
	appendFile("File_Xtensa", values.Xtensa, values.KindFirmware)
	appendFile("File_RISC-V", values.RISCV, values.KindFirmware)

	if len(files) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	created, err := c.service.UploadBuild(
		values.Slug(slug),
		values.OS(platform),
		buildUploadRequest.BasicInfo,
		files,
		buildUploadRequest.Metadata,
		buildsrv.UploadOptions{
			Actor:          actor(r),
			Overwrite:      r.Form.Get("overwrite") == "true",
			ExpectedSHA256: r.Form.Get("Sha256"),
		},
	)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "build number already exists" {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	if entity, err := c.appService.GetBySlug(values.Slug(slug)); err == nil && entity.TeamsWebhookURL != "" {
		title := "Build " + created.Number + " uploaded"
		text := slug + " " + platform + " " + created.Version
		url := entity.TeamsWebhookURL
		go func() { _ = teams.Post(context.Background(), url, title, text) }()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (c *buildController) HandleArtifactDownload(w http.ResponseWriter, r *http.Request) {
	checksum := chi.URLParam(r, "checksum")
	if !storage.ValidChecksum(checksum) {
		http.Error(w, "Invalid checksum", http.StatusBadRequest)
		return
	}

	filename := checksum
	if artifact, err := c.service.GetArtifactByChecksum(checksum); err == nil {
		filename = safeAttachmentName(artifact.Filename)
	}

	filePath := storage.ConvertChecksumToPath(checksum)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+safeAttachmentName(filename)+`"`)
	_ = c.service.IncrementDownload(checksum)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, file); err != nil {
		return
	}
}

func (c *buildController) HandlePublishBuild(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	platform := chi.URLParam(r, "platform")
	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !c.canEdit(r, slug) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	currentPlatform := session.NewCurrentPlatformData(currentApp, values.OS(platform))
	if currentPlatform.PlatformID == 0 {
		http.NotFound(w, r)
		return
	}
	if err = c.service.PublishBuild(currentPlatform.PlatformID, number, actor(r)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/platforms/"+platform+"/builds/"+strconv.Itoa(number), http.StatusSeeOther)
}

func (c *buildController) HandleDeleteBuild(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	platform := chi.URLParam(r, "platform")
	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !c.canEdit(r, slug) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	currentPlatform := session.NewCurrentPlatformData(currentApp, values.OS(platform))
	if currentPlatform.PlatformID == 0 {
		http.NotFound(w, r)
		return
	}
	if err = c.service.DeleteBuild(currentPlatform.PlatformID, number, actor(r)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/platforms/"+platform+"/builds", http.StatusSeeOther)
}

func safeAttachmentName(name string) string {
	base := filepath.Base(strings.ReplaceAll(name, `"`, ""))
	if base == "" || base == "." || base == ".." {
		return "download"
	}
	return base
}

func actor(r *http.Request) string {
	if value, ok := r.Context().Value("actor").(string); ok && value != "" {
		return value
	}
	if value, ok := r.Context().Value("username").(string); ok {
		return value
	}
	return ""
}
