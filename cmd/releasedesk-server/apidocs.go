package main

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec []byte

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(openAPISpec)
}

func serveAPIDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><title>ReleaseDesk API</title>
<style>body{font-family:sans-serif;margin:2rem;max-width:52rem} code{background:#f4f4f5;padding:.1rem .3rem}</style></head><body>
<h1>ReleaseDesk API</h1>
<p>OpenAPI document: <a href="/api/openapi.yaml">/api/openapi.yaml</a>. Authenticate with <code>Authorization: Bearer</code> or <code>X-Auth</code> using a per-app CI token.</p>
<ul>
<li>GET /api/v1/apps</li>
<li>POST /api/v1/apps/{slug}/releases</li>
<li>GET /api/v1/apps/{slug}/releases/{version}</li>
<li>GET /api/v1/apps/{slug}/releases/{version}/files</li>
<li>GET /api/v1/apps/{slug}/releases/{version}/bundle.zip</li>
<li>PUT /api/v1/apps/{slug}/releases/{version}/notes</li>
<li>PUT /api/v1/apps/{slug}/releases/{version}/test-methods</li>
<li>GET /api/v1/apps/{slug}/feed.json</li>
<li>POST /api/apps/{slug}/platforms/{platform}/builds</li>
<li>GET/PUT /api/v1/apps/{slug}/members (console session)</li>
<li>POST /api/v1/apps/{slug}/tokens (console session)</li>
<li>PUT /api/v1/apps/{slug}/settings and POST .../webhooks/test</li>
</ul>
<p>See docs/CI.md and docs/INFRASTRUCTURE.md in the repository.</p>
</body></html>`))
}
