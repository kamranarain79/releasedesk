# Python tool + Filebrowser

Production flashing stays on [Filebrowser](https://github.com/filebrowser/filebrowser). ReleaseDesk stores firmware and application artifacts, release notes, and QA test methods. Your existing Python tool can pull a release from ReleaseDesk (same idea as listing/downloading through Filebrowser) and then copy files into Filebrowser for the factory.

Auth (any one of these):

- `Authorization: Bearer <token>`
- `X-Auth: <token>` (Filebrowser-style header)
- HTTP Basic, password = token
- Query `?token=`

Use the per-app CI token from the app Settings page. The instance token is generated on first start and shown under Administration until you dismiss it. Do not put the token in the query string.

## Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/apps` | Apps on this instance |
| GET | `/api/v1/apps/{slug}/releases/{version}` | Bundle: notes, QA tests, files |
| GET | `/api/v1/apps/{slug}/releases/{version}/files` | Filebrowser-like directory listing |
| GET | `/api/v1/apps/{slug}/releases/{version}/files/{name}` | Download one file by name |
| GET | `/api/v1/apps/{slug}/releases/{version}/notes` | Release notes + changelog |
| GET | `/api/v1/apps/{slug}/releases/{version}/test-methods` | QA test plan |
| GET | `/api/v1/apps/{slug}/channels/{channel}/latest` | Redirect to latest version on that channel |
| GET | `/api/v1/files/{sha256}` | Download by checksum |
| POST | `/api/apps/{slug}/platforms/{platform}/builds` | Upload a build (multipart) |

`{platform}` is `ESP32`, `STM32`, `Linux`, `Android`, …  
ESP32 form fields: `File_firmware`, `File_bootloader`, `File_partition`, `File_application`.  
Linux RAUC: `File_rauc` (`.raucb`) or a normal package as `File`.

## Example

```python
import requests

BASE = "http://localhost:8080"
TOKEN = "paste-the-token-from-administration"
HEADERS = {"Authorization": f"Bearer {TOKEN}"}

slug, version = "rw", "1.0.0"
bundle = requests.get(
    f"{BASE}/api/v1/apps/{slug}/releases/{version}",
    headers=HEADERS,
    timeout=30,
).json()

print(bundle["notes"])
for method in bundle["testMethods"]:
    print(method["title"], method["steps"], method["expectedResult"])

listing = requests.get(
    f"{BASE}/api/v1/apps/{slug}/releases/{version}/files",
    headers=HEADERS,
    timeout=30,
).json()

for item in listing["items"]:
    raw = requests.get(BASE + item["url"], headers=HEADERS, timeout=60)
    raw.raise_for_status()
    with open(item["name"], "wb") as f:
        f.write(raw.content)
    # then upload into Filebrowser as you already do for production flash
```

Files on a release are:

1. Builds linked to that release, plus
2. Platform builds whose **version** matches the release version (ESP32 firmware + Linux RAUC + STM32 HEX can share `1.0.0`).
