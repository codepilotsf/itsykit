package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"flea/internal/contact"
	"flea/internal/devaccess"
	"flea/internal/editor"
	"flea/internal/site"
)

// Exercise the real dev admin/store/public pipeline as a different LAN device.
// Only email delivery is captured; authentication and publication are unchanged.
func TestLANDevWorkflow(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	root := filepath.Join(t.TempDir(), "site")
	if err := createSite(root, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := site.AddAdmin(root, "owner@example.test"); err != nil {
		t.Fatal(err)
	}
	enableTestContact(t, root)
	var photo bytes.Buffer
	if err := png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	video, err := os.ReadFile("../video/testdata/valid.mp4")
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"site/images/lan.png": photo.Bytes(),
		"site/images/lan.mp4": video,
		"site/lan-media.html": []byte(`<h1 snip>LAN media</h1><img snip src="/images/lan.png" alt="Seed"><video snip controls src="/images/lan.mp4"></video>`),
	} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	p, err := site.LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	mail := make(previewInbox, 4)
	admin := devAdmin(root, "LAN test", public, nil, mail)
	stopApps := public.StartApps(context.Background(), mail, &contact.Budget{})
	defer stopApps()
	h := devaccess.WrapLAN(admin, []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	request := func(handler http.Handler, method, path string, values url.Values, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "http://asahi:3000"+path, strings.NewReader(values.Encode()))
		r.RemoteAddr = "100.79.67.33:54321"
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("X-CSRF-Token", values.Get("csrf"))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	const origin = "http://asahi:3000"
	req := func(method, path string, values url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
		return request(h, method, path, values, cookie, origin)
	}
	cookie := func(w *httptest.ResponseRecorder, name string) *http.Cookie {
		t.Helper()
		for _, c := range w.Result().Cookies() {
			if c.Name == name {
				return c
			}
		}
		t.Fatalf("missing %s cookie: %d", name, w.Code)
		return nil
	}
	if w := request(admin, "GET", "/admin", nil, nil, origin); w.Code != 403 {
		t.Fatal("LAN enabled without middleware", w.Code)
	}
	if w := request(public, "POST", "/contact", nil, nil, origin); w.Code != 403 {
		t.Fatal("contact LAN enabled by default", w.Code)
	}
	productionAdmin := devaccess.WrapLAN(editor.New(editor.Options{Emails: []string{"owner@example.test"}}), []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	if w := request(productionAdmin, "GET", "/admin", nil, nil, origin); w.Code != 403 || !strings.Contains(w.Body.String(), "HTTPS") {
		t.Fatal("LAN context bypassed production HTTPS")
	}
	productionForm := devaccess.WrapLAN(site.NewProductionServer(p, nil), []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	if w := request(productionForm, "POST", "/contact", nil, nil, origin); w.Code != 403 {
		t.Fatal("LAN context bypassed production contact HTTPS")
	}
	w := req("GET", "/admin", nil, nil)
	login := cookie(w, "flea_dev_login")
	w = req("POST", "/admin/code", url.Values{"csrf": {login.Value}, "email": {"owner@example.test"}}, login)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var code string
	select {
	case message := <-mail:
		code = regexp.MustCompile(`\b[0-9]{6}\b`).FindString(message.Text)
	case <-time.After(2 * time.Second):
		t.Fatal("no sign-in preview")
	}
	w = req("POST", "/admin/verify", url.Values{"csrf": {login.Value}, "code": {code}}, login)
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	session := cookie(w, "flea_dev_session")
	if session.Secure || !session.HttpOnly || session.SameSite != http.SameSiteStrictMode {
		t.Fatal("unexpected dev cookie policy")
	}
	w = req("GET", "/", nil, session)
	csrfMatch := regexp.MustCompile(`name="csrf-token" content="([^"]+)"`).FindStringSubmatch(w.Body.String())
	if len(csrfMatch) != 2 || !strings.Contains(w.Body.String(), "flea-owner-root") {
		t.Fatal("LAN owner shell missing", w.Code)
	}
	csrf := csrfMatch[1]
	metadata := func(route string) editor.PageEdit {
		t.Helper()
		w := req("GET", "/admin/editor-context?route="+url.QueryEscape(route), url.Values{"csrf": {csrf}}, session)
		var data struct {
			Page editor.PageEdit `json:"page"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil || w.Code != 200 {
			t.Fatal("editor context", w.Code, err)
		}
		return data.Page
	}
	edit := metadata("/")
	edit.Title = "Saved from another device"
	data, _ := json.Marshal(edit)
	values := url.Values{"csrf": {csrf}, "edit": {string(data)}}
	if w := request(h, "POST", "/admin/page", values, session, "http://evil.test"); w.Code != 403 {
		t.Fatal("foreign origin accepted", w.Code)
	}
	if w := req("POST", "/admin/page", url.Values{"csrf": {"wrong"}, "edit": {string(data)}}, session); w.Code != 403 {
		t.Fatal("bad CSRF accepted", w.Code)
	}
	if w := req("POST", "/admin/page", values, session); w.Code != 200 {
		t.Fatal("page save", w.Code, w.Body.String())
	}
	if w := req("POST", "/admin/page", values, session); w.Code != 409 {
		t.Fatal("stale save accepted", w.Code)
	}
	if w := req("GET", "/", nil, nil); !strings.Contains(w.Body.String(), edit.Title) || strings.Contains(w.Body.String(), "flea-owner-root") {
		t.Fatal("anonymous publication failed")
	}
	// The same authenticated LAN endpoint must support bounded multipart media.
	mediaEdit := metadata("/lan-media")
	var imageID, videoID string
	for id, entry := range p.Content.Pages["/lan-media"].Snips {
		switch entry.Type {
		case "image":
			imageID = id
		case "video":
			videoID = id
		}
	}
	if imageID == "" || videoID == "" {
		t.Fatal("media seeds not extracted")
	}
	mediaEdit.Media = []editor.MediaEdit{{ID: imageID, Name: "LAN photo", Alt: "Uploaded from LAN"}, {ID: videoID, Name: "LAN video"}}
	data, _ = json.Marshal(mediaEdit)
	var upload bytes.Buffer
	parts := multipart.NewWriter(&upload)
	if err := parts.WriteField("csrf", csrf); err != nil {
		t.Fatal(err)
	}
	if err := parts.WriteField("edit", string(data)); err != nil {
		t.Fatal(err)
	}
	for _, file := range []struct {
		id, name string
		data     []byte
	}{{imageID, "photo.png", photo.Bytes()}, {videoID, "video.mp4", video}} {
		part, err := parts.CreateFormFile("media-"+file.id, file.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = part.Write(file.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := parts.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", origin+"/admin/page", bytes.NewReader(upload.Bytes()))
	r.RemoteAddr = "100.79.67.33:54321"
	r.Header.Set("Origin", origin)
	r.Header.Set("Content-Type", parts.FormDataContentType())
	r.Header.Set("X-CSRF-Token", csrf)
	r.AddCookie(session)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("LAN media upload", w.Code, w.Body.String())
	}
	if w := req("GET", "/lan-media", nil, nil); !strings.Contains(w.Body.String(), "Uploaded from LAN") {
		t.Fatal("uploaded media not published")
	}
	// Collections retain the normal version checks and publication semantics.
	team := metadata("/team")
	items := team.Multisnips["team"]
	if len(items) < 2 {
		t.Fatal("missing starter collection")
	}
	items[0], items[1] = items[1], items[0]
	items[0].Fields["name"] = "LAN team member"
	data, _ = json.Marshal(team)
	if w := req("POST", "/admin/page", url.Values{"csrf": {csrf}, "edit": {string(data)}}, session); w.Code != 200 {
		t.Fatal("collection save", w.Code, w.Body.String())
	}
	if w := req("GET", "/team", nil, nil); !strings.Contains(w.Body.String(), "LAN team member") {
		t.Fatal("collection not published")
	}
	for _, method := range []string{"GET", "POST"} {
		if w := req(method, "/admin/theme", url.Values{"csrf": {csrf}}, session); w.Code != 404 {
			t.Fatal("removed Theme endpoint", method, w.Code)
		}
	}
	if w := req("POST", "/admin/activity", url.Values{"csrf": {csrf}}, session); w.Code != 204 {
		t.Fatal("activity", w.Code)
	}
	if submitTestContact(t, h, origin, "192.168.1.9:54321", "LAN contact test") {
		select {
		case message := <-mail:
			if message.To != "contact@example.test" {
				t.Fatal("wrong recipient")
			}
		case <-time.After(time.Second):
			t.Fatal("missing contact preview")
		}
	}
	before := req("GET", "/_flea/reload", nil, nil).Body.String()
	if err := os.WriteFile(filepath.Join(root, "site/test.html"), []byte("<p>LAN reload test</p>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	if w := req("GET", "/_flea/reload", nil, nil); w.Code != 200 || w.Body.String() == before {
		t.Fatal("reload revision did not change")
	}
	if w := req("POST", "/", nil, nil); w.Code != 405 || !strings.Contains(w.Body.String(), "Check the form") {
		t.Fatal("LAN diagnostic unavailable", w.Code)
	}
	for _, path := range []string{"/.flea.config.json", "/flea/runtime/sessions.json", "/mailers/sign-in.email.html"} {
		if w := req("GET", path, nil, nil); w.Code != 404 {
			t.Fatalf("private path %s: %d", path, w.Code)
		}
	}
	if w := req("POST", "/admin/logout", url.Values{"csrf": {csrf}}, session); w.Code != 303 {
		t.Fatal("logout", w.Code)
	}
	if w := req("GET", "/", nil, session); strings.Contains(w.Body.String(), "flea-owner-root") {
		t.Fatal("logged-out session still works")
	}
}
