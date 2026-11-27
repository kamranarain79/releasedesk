package values

// ArtifactKind is the role of a file inside a build (firmware image vs partition table, etc.).
type ArtifactKind string

const (
	KindFirmware    ArtifactKind = "firmware"
	KindBootloader  ArtifactKind = "bootloader"
	KindPartition   ArtifactKind = "partition"
	KindApplication ArtifactKind = "application"
	KindHex         ArtifactKind = "hex"
	KindRAUC        ArtifactKind = "rauc"
)

func ParseArtifactKind(name string) ArtifactKind {
	switch ArtifactKind(name) {
	case KindBootloader, KindPartition, KindApplication, KindHex, KindRAUC, KindFirmware:
		return ArtifactKind(name)
	default:
		return KindFirmware
	}
}
