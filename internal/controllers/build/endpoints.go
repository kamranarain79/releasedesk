package build

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	domainbuild "github.com/brewbits-co/releasedesk/internal/domains/build"
	buildsrv "github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/fields"
	"github.com/brewbits-co/releasedesk/pkg/storage"
	"github.com/brewbits-co/releasedesk/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (c *buildController) HandleBuildUpload(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(512 << 20)

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

	_, err = c.service.UploadBuild(
		values.Slug(slug),
		values.OS(platform),
		buildUploadRequest.BasicInfo,
		files,
		buildUploadRequest.Metadata,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *buildController) HandleArtifactDownload(w http.ResponseWriter, r *http.Request) {
	checksum := chi.URLParam(r, "checksum")
	if len(checksum) < 2 {
		http.Error(w, "Invalid checksum", http.StatusBadRequest)
		return
	}

	filename := checksum
	if artifact, err := c.service.GetArtifactByChecksum(checksum); err == nil {
		filename = artifact.Filename
	}

	subFolder := checksum[:2]
	filePath := filepath.Join(storage.StorageDir, subFolder, checksum)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "Error serving file", http.StatusInternalServerError)
		return
	}
}
