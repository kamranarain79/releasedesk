package sql

import (
	"database/sql"
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"xorm.io/xorm"
)

// NewBuildRepository is the constructor for buildRepository
func NewBuildRepository(engine *xorm.Engine) build.BuildRepository {
	return &buildRepository{engine: engine}
}

// buildRepository is the implementation of build.BuildRepository
type buildRepository struct {
	engine *xorm.Engine
}

func (r *buildRepository) Save(buildEntity *build.Build) error {
	session := r.engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	if _, err := session.Insert(buildEntity); err != nil {
		_ = session.Rollback()
		return err
	}

	for i := range buildEntity.Artifacts {
		buildEntity.Artifacts[i].BuildID = buildEntity.ID
		if _, err := session.Insert(&buildEntity.Artifacts[i]); err != nil {
			_ = session.Rollback()
			return err
		}
	}

	for key, value := range buildEntity.Metadata {
		metadata := build.BuildMetadata{
			BuildID: buildEntity.ID,
			Key:     key,
			Value:   value,
		}
		if _, err := session.Insert(&metadata); err != nil {
			_ = session.Rollback()
			return err
		}
	}

	if err := session.Commit(); err != nil {
		_ = session.Rollback()
		return err
	}

	return nil
}

func (r *buildRepository) ExistsByNumber(platformID int, number string) (bool, error) {
	return r.engine.Where("platform_id = ? AND number = ?", platformID, number).Exist(&build.Build{})
}

func (r *buildRepository) DeleteByPlatformAndNumber(platformID int, number string) error {
	var existing build.Build
	has, err := r.engine.Where("platform_id = ? AND number = ?", platformID, number).Get(&existing)
	if err != nil || !has {
		return err
	}
	session := r.engine.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		return err
	}
	if _, err = session.Where("build_id = ?", existing.ID).Delete(&build.Artifact{}); err != nil {
		_ = session.Rollback()
		return err
	}
	if _, err = session.Where("build_id = ?", existing.ID).Delete(&build.BuildMetadata{}); err != nil {
		_ = session.Rollback()
		return err
	}
	if _, err = session.ID(existing.ID).Delete(&build.Build{}); err != nil {
		_ = session.Rollback()
		return err
	}
	return session.Commit()
}

func (r *buildRepository) FindByPlatformID(platformID int) ([]build.BasicInfo, error) {
	// Execute the database query
	var builds []build.BasicInfo
	err := r.engine.Desc("created_at").Table("build").Find(&builds, &build.BasicInfo{PlatformID: platformID})
	if err != nil {
		return nil, err // Return an error if the query fails
	}

	// Format timestamps for each build
	for i := range builds {
		builds[i].FormatAuditable()
	}

	return builds, nil
}

func (r *buildRepository) FindByPlatformIDAndVersion(platformID int, version string) ([]build.BasicInfo, error) {
	var builds []build.BasicInfo
	err := r.engine.Desc("created_at").Table("build").Where("platform_id = ? AND version = ?", platformID, version).Find(&builds)
	if err != nil {
		return nil, err
	}
	for i := range builds {
		builds[i].FormatAuditable()
	}
	return builds, nil
}

func (r *buildRepository) GetByID(id int) (build.Build, error) {
	var buildDetails build.Build
	exist, err := r.engine.ID(id).Get(&buildDetails)
	if err != nil {
		return build.Build{}, err
	}
	if !exist {
		return build.Build{}, sql.ErrNoRows
	}
	buildDetails.Auditable.FormatAuditable()

	var artifacts []build.Artifact
	if err = r.engine.Where("build_id = ?", buildDetails.ID).Find(&artifacts); err != nil {
		return build.Build{}, err
	}
	buildDetails.Artifacts = artifacts
	return buildDetails, nil
}

func (r *buildRepository) SetStatus(id int, status string) error {
	_, err := r.engine.Exec("UPDATE build SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *buildRepository) CountArtifactsByChecksum(checksum string) (int64, error) {
	return r.engine.Where("sha256 = ?", checksum).Count(&build.Artifact{})
}

func (r *buildRepository) DeleteByID(id int) ([]string, error) {
	var artifacts []build.Artifact
	if err := r.engine.Where("build_id = ?", id).Find(&artifacts); err != nil {
		return nil, err
	}
	checksums := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Sha256 != "" {
			checksums = append(checksums, artifact.Sha256)
		}
	}
	session := r.engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	if _, err := session.Exec("DELETE FROM linked_builds WHERE build_id = ?", id); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if _, err := session.Where("build_id = ?", id).Delete(&build.Artifact{}); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if _, err := session.Where("build_id = ?", id).Delete(&build.BuildMetadata{}); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if _, err := session.Exec("DELETE FROM build WHERE id = ?", id); err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if err := session.Commit(); err != nil {
		return nil, err
	}
	return checksums, nil
}

func (r *buildRepository) GetArtifactByChecksum(checksum string) (build.Artifact, error) {
	var artifact build.Artifact
	has, err := r.engine.Where("sha256 = ?", checksum).Get(&artifact)
	if err != nil {
		return build.Artifact{}, err
	}
	if !has {
		return build.Artifact{}, sql.ErrNoRows
	}
	return artifact, nil
}

func (r *buildRepository) GetByPlatformIDAndNumber(platformID int, number int) (build.Build, error) {
	var buildDetails build.Build

	exist, err := r.engine.Where("platform_id = ? AND number = ?", platformID, number).Get(&buildDetails)
	if err != nil {
		return build.Build{}, err
	}
	if !exist {
		return build.Build{}, sql.ErrNoRows
	}

	buildDetails.Auditable.FormatAuditable()

	// Fetch build artifacts
	var artifacts []build.Artifact
	err = r.engine.Where("build_id = ?", buildDetails.ID).Find(&artifacts)
	if err != nil {
		return build.Build{}, err
	}
	buildDetails.Artifacts = artifacts

	// Fetch build metadata
	var metadataItems []build.BuildMetadata
	err = r.engine.Where("build_id = ?", buildDetails.ID).Find(&metadataItems)
	if err != nil {
		return build.Build{}, err
	}

	metadata := make(map[string]string)
	for _, item := range metadataItems {
		metadata[item.Key] = item.Value
	}
	buildDetails.Metadata = metadata

	return buildDetails, nil
}
