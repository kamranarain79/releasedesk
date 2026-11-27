<h1 align="center" style="border-bottom: none">
    <a href="https://signoz.io" target="_blank">
        <img alt="SigNoz" src="https://github.com/user-attachments/assets/17d97874-a835-4159-9d05-3a5a9eb53efd" width="80" height="80">
    </a>
    <br>ReleaseDesk
</h1>
<p align="center">An open-source & self-hostable app distribution tool.</p>

## About the Project

**ReleaseDesk** is an open-source & self-hostable alternative to traditional App Stores / AppCenter / TestFlight / etc.

It helps you manage app releases, write meaningful release notes, and publish them across multiple channels (**beta**, **stable**, or custom ones).

With **ReleaseDesk**, you can:

- 📦 **Distribute your apps** both internally and externally through a customizable **Download Portal**
- 🔐 **Control access** using **distribution groups**
- 🚀 Release firmware **and** application binaries for **multiple platforms**:
    - 🪟 Windows (`.appx`, `.appxbundle`, `.appxupload`, `.msix`, `.msixbundle`, `.msixupload`, `.exe`, `.zip`, `.msi`)
    - 🍎 macOS (`.zip`, `.app.zip`, `.dmg`, `.pkg`)
    - 🐧 Linux (`.AppImage`, `.deb`, `.rpm`, `.tgz`, `.gz`, `.snap`, `.flatpak`, `.raucb`, `.img`, `.wic`)
    - 📱 iOS (`.ipa`)
    - 🤖 Android (`.apk`, `.aab`)
    - 🔌 ESP32 (firmware `.bin`, bootloader, partition table)
    - 🔌 STM32 (`.bin` / Intel HEX)
- 📝 Keep **release notes** and **QA test methods** on each version
- 🏭 Production flashing still uses [Filebrowser](https://github.com/filebrowser/filebrowser); the Python tool pulls files/notes/tests from `/api/v1` (see [docs/PYTHON_TOOL.md](docs/PYTHON_TOOL.md))
- ⚙️ **Automate your CI/CD pipeline** with release workflows that:
    - 📣 Trigger announcements (email notifications, webhooks, etc.)
    - 🔄 Integrate with your app for **in-app updates**
    - 📡 Generate update feeds in formats like **JSON**, **XML**, **Squirrel**, **Sparkle**, and more

> 🚧 This project is in its early stages.  
Share your ideas or questions on [Discussions](https://github.com/brewbits-co/releasedesk/discussions) page.

## Local development

**Docker** (pins Go 1.23 and Node 22, embeds a matching UI bundle, and starts PostgreSQL with the app):

```bash
docker compose up --build
```

Or with a host toolchain (Go **1.23+** and Node **22**):

```bash
make build && make run
```

Open **http://localhost:8080/login** (default: `admin` / `admin`).  
Public portal: **http://localhost:9090**.

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for Air, troubleshooting, and which console pages are implemented vs placeholders.  
Hosted product documentation: [releasedesk.dev/docs](https://releasedesk.dev/docs/).
