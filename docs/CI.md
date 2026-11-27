# CI publishing

1. In the console, create the app, finish the setup guide (channels such as Stable), and add ESP32, STM32, and Linux.
2. On **Settings**, add people as app admin or app user, paste the Teams Incoming Webhook, and mint a CI token. Copy it once.
3. In the firmware GitLab project, set masked protected variables `RELEASEDESK_URL` (https), `RELEASEDESK_APP_SLUG`, `RELEASEDESK_APP_TOKEN`.
4. Include [examples/gitlab/push-to-releasedesk.yml](../examples/gitlab/push-to-releasedesk.yml).

The job creates the release, then uploads artifacts that share the same version so the factory listing unions them. `overwrite=true` replaces the same build number. Without it, a retry returns 409.

Optional `Sha256` form field is checked when a single file is uploaded.

Factory download:

```bash
curl -fsS -H "Authorization: Bearer $RELEASEDESK_APP_TOKEN" \
  "$RELEASEDESK_URL/api/v1/apps/$RELEASEDESK_APP_SLUG/releases/$CI_COMMIT_TAG/bundle.zip" -o release.zip
```

Or list files (Filebrowser-style `items`) and copy them into Filebrowser. Teams receives a card; a Teams outage does not fail the job.

This repository's own pipeline is [.gitlab-ci.yml](../.gitlab-ci.yml) (web, Go test, image build). It does not deploy automatically.
