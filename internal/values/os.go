package values

type OS string

const (
	Android OS = "Android"
	IOS     OS = "iOS"
	Windows OS = "Windows"
	MacOS   OS = "macOS"
	Linux   OS = "Linux"
	ESP32   OS = "ESP32"
	STM32   OS = "STM32"
)

func (o OS) IsFirmware() bool {
	return o == ESP32 || o == STM32
}

func AllowedExtensions(os OS) []string {
	switch os {
	case Windows:
		return []string{".appx", ".appxbundle", ".appxupload", ".msix", ".msixbundle", ".msixupload", ".exe", ".zip", ".msi"}
	case MacOS:
		return []string{".zip", ".platform.zip", ".dmg", ".pkg"}
	case Linux:
		return []string{".AppImage", ".deb", ".rpm", ".tgz", ".gz", ".snap", ".flatpak", ".raucb", ".img", ".wic", ".squashfs"}
	case Android:
		return []string{".apk", ".aab"}
	case IOS:
		return []string{".ipa"}
	case ESP32:
		return []string{".bin", ".elf", ".csv", ".json"}
	case STM32:
		return []string{".bin", ".hex", ".elf", ".srec"}
	default:
		return nil
	}
}
