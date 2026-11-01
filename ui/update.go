package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	latestReleaseAPI = "https://api.github.com/repos/lazymule/envui/releases/latest"
	prefCheckUpdates = "checkUpdates"
	prefLastCheck    = "lastUpdateCheck"
	updateInterval   = 24 * time.Hour
)

type releaseInfo struct {
	Tag string `json:"tag_name"`
	URL string `json:"html_url"`
}

// fetchLatestRelease asks GitHub for the newest published release.
func fetchLatestRelease() (releaseInfo, error) {
	var r releaseInfo
	req, err := http.NewRequest(http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return r, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "envui/"+Version)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return r, fmt.Errorf("GitHub: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return r, err
	}
	if r.Tag == "" {
		return r, errors.New("GitHub: no release found")
	}
	return r, nil
}

// parseVersion turns "v1.2.3" / "1.2" into numbers; pre-release suffixes
// ("-beta") are ignored.
func parseVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+ "); i >= 0 {
		v = v[:i]
	}
	var out []int
	for _, p := range strings.Split(v, ".") {
		n, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		out = append(out, n)
	}
	return out
}

// isNewer reports whether version a is newer than b.
func isNewer(a, b string) bool {
	va, vb := parseVersion(a), parseVersion(b)
	for i := 0; i < len(va) || i < len(vb); i++ {
		var x, y int
		if i < len(va) {
			x = va[i]
		}
		if i < len(vb) {
			y = vb[i]
		}
		if x != y {
			return x > y
		}
	}
	return false
}

// autoCheckForUpdates runs a background check at most once a day, if the
// user hasn't turned it off in Settings.
func (a *App) autoCheckForUpdates() {
	if !a.prefs.BoolWithFallback(prefCheckUpdates, true) {
		return
	}
	if last, err := strconv.ParseInt(a.prefs.String(prefLastCheck), 10, 64); err == nil &&
		time.Since(time.Unix(last, 0)) < updateInterval {
		// checked recently; still show a notice if we already know of one
		if tag, link := a.prefs.String("latestTag"), a.prefs.String("latestURL"); tag != "" && isNewer(tag, Version) {
			a.showUpdateNotice(releaseInfo{Tag: tag, URL: link})
		}
		return
	}
	go func() {
		r, err := fetchLatestRelease()
		if err != nil {
			return // offline etc.: try again next start
		}
		a.prefs.SetString(prefLastCheck, strconv.FormatInt(time.Now().Unix(), 10))
		a.prefs.SetString("latestTag", r.Tag)
		a.prefs.SetString("latestURL", r.URL)
		if isNewer(r.Tag, Version) {
			a.showUpdateNotice(r)
		}
	}()
}

// checkForUpdatesNow is the manual check from the About dialog.
func (a *App) checkForUpdatesNow() {
	progress := dialog.NewCustomWithoutButtons(T("update.title"), widget.NewProgressBarInfinite(), a.window)
	progress.Show()
	go func() {
		r, err := fetchLatestRelease()
		progress.Hide()
		if err != nil {
			a.showError(errors.New(T("update.failed", err)))
			return
		}
		a.prefs.SetString(prefLastCheck, strconv.FormatInt(time.Now().Unix(), 10))
		if !isNewer(r.Tag, Version) {
			a.showInfo(T("update.title"), T("update.upToDate", Version))
			return
		}
		a.showUpdateNotice(r)
		a.showConfirm(T("update.title"), T("update.available", r.Tag, Version), T("update.open"), func() {
			a.openURL(r.URL)
		})
	}()
}

func (a *App) showUpdateNotice(r releaseInfo) {
	if a.updateLink == nil {
		return
	}
	a.updateLink.SetText(T("update.notice", r.Tag))
	if u, err := url.Parse(r.URL); err == nil && r.URL != "" {
		a.updateLink.SetURL(u)
	}
	a.updateLink.Show()
}

func (a *App) openURL(raw string) {
	if u, err := url.Parse(raw); err == nil {
		_ = a.fyneApp.OpenURL(u)
	}
}

// AfterStart runs background work once the window is set up.
func (a *App) AfterStart() {
	a.autoCheckForUpdates()
}
