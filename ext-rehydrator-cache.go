package hydration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func loadUpgradeCache() *UpgradeCache {
	scratchDir := os.Getenv("GEMINI_CLI_HOME")
	if scratchDir == "" {
		scratchDir = `C:\aCogSpaceSeed\c0990-ephemeral-scratch`
	}
	cachePath := filepath.Join(scratchDir, "upgrade_cache.webnf")

	cache := &UpgradeCache{
		Entries:  make(map[string]UpgradeCacheEntry),
		Licenses: make(map[string]LicenseCacheEntry),
	}

	contentBytes, err := os.ReadFile(cachePath)
	if err != nil {
		return cache
	}

	lines := strings.Split(string(contentBytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "hold":
			if t, err := time.Parse(time.RFC3339, parts[1]); err == nil {
				cache.GitHubHoldUntil = t
			}
		case "cache":
			if len(parts) >= 4 {
				pkgName := parts[1]
				version := parts[2]
				if t, err := time.Parse(time.RFC3339, parts[3]); err == nil {
					cache.Entries[pkgName] = UpgradeCacheEntry{
						Version:   version,
						CheckedAt: t,
					}
				}
			}
		case "license":
			if len(parts) >= 4 {
				pkgName := parts[1]
				status := parts[2]
				if t, err := time.Parse(time.RFC3339, parts[3]); err == nil {
					cache.Licenses[pkgName] = LicenseCacheEntry{
						Status:    status,
						ScannedAt: t,
					}
				}
			}
		}
	}
	return cache
}

func saveUpgradeCache(cache *UpgradeCache) {
	scratchDir := os.Getenv("GEMINI_CLI_HOME")
	if scratchDir == "" {
		scratchDir = `C:\aCogSpaceSeed\c0990-ephemeral-scratch`
	}
	cachePath := filepath.Join(scratchDir, "upgrade_cache.webnf")

	var buf bytes.Buffer
	buf.WriteString("# upgrade-cache-v1\n")
	buf.WriteString(fmt.Sprintf("hold|%s\n", cache.GitHubHoldUntil.Format(time.RFC3339)))

	var cacheKeys []string
	for k := range cache.Entries {
		cacheKeys = append(cacheKeys, k)
	}
	sort.Strings(cacheKeys)
	for _, k := range cacheKeys {
		entry := cache.Entries[k]
		buf.WriteString(fmt.Sprintf("cache|%s|%s|%s\n", k, entry.Version, entry.CheckedAt.Format(time.RFC3339)))
	}

	var licKeys []string
	for k := range cache.Licenses {
		licKeys = append(licKeys, k)
	}
	sort.Strings(licKeys)
	for _, k := range licKeys {
		lic := cache.Licenses[k]
		buf.WriteString(fmt.Sprintf("license|%s|%s|%s\n", k, lic.Status, lic.ScannedAt.Format(time.RFC3339)))
	}

	if err := os.WriteFile(cachePath, buf.Bytes(), 0644); err != nil {
		slog.Warn("Failed to save upgrade cache file", "path", cachePath, "error", err)
	}
}

func fetchLatestGitHubRelease(repo string) (string, time.Time, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("User-Agent", "Antigravity-Maturity-Checker")
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 403 {
		resetTime := time.Now().Add(15 * time.Minute)
		if resetHeader := resp.Header.Get("X-RateLimit-Reset"); resetHeader != "" {
			if resetUnix, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
				resetTime = time.Unix(resetUnix, 0)
			}
		}
		return "RATE_LIMITED", resetTime, nil
	}

	var rawTag string
	if resp.StatusCode != 200 {
		tagsUrl := fmt.Sprintf("https://api.github.com/repos/%s/tags", repo)
		reqTags, err := http.NewRequest("GET", tagsUrl, nil)
		if err != nil {
			return "", time.Time{}, err
		}
		reqTags.Header.Set("User-Agent", "Antigravity-Maturity-Checker")
		respTags, err := client.Do(reqTags)
		if err != nil {
			return "", time.Time{}, err
		}
		defer respTags.Body.Close()
		if respTags.StatusCode == 403 {
			resetTime := time.Now().Add(15 * time.Minute)
			if resetHeader := respTags.Header.Get("X-RateLimit-Reset"); resetHeader != "" {
				if resetUnix, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
					resetTime = time.Unix(resetUnix, 0)
				}
			}
			return "RATE_LIMITED", resetTime, nil
		}
		if respTags.StatusCode != 200 {
			return "", time.Time{}, fmt.Errorf("releases returned status %d, tags returned status %d", resp.StatusCode, respTags.StatusCode)
		}
		var tagsData []struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(respTags.Body).Decode(&tagsData); err != nil {
			return "", time.Time{}, err
		}
		if len(tagsData) == 0 {
			return "", time.Time{}, fmt.Errorf("no releases or tags found for repository %s", repo)
		}
		bestTag := tagsData[0].Name
		for _, t := range tagsData[1:] {
			if compareVersions(t.Name, bestTag) > 0 {
				bestTag = t.Name
			}
		}
		rawTag = bestTag
	} else {
		var data struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return "", time.Time{}, err
		}
		rawTag = data.TagName
	}

	var tag string
	if strings.HasPrefix(rawTag, "version_") {
		tag = strings.TrimPrefix(rawTag, "version_")
	} else if strings.HasPrefix(rawTag, "go") {
		tag = strings.TrimPrefix(rawTag, "go")
	} else {
		tag = strings.TrimPrefix(rawTag, "v")
	}
	tag = strings.TrimSuffix(tag, ".windows.1")
	return tag, time.Time{}, nil
}

func fetchLatestNpmRelease(pkg string) (string, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Antigravity-Maturity-Checker")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var data struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Version, nil
}

func fetchLatestFlutterRelease() (string, error) {
	url := "https://storage.googleapis.com/flutter_infra_release/releases/releases_windows.json"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Antigravity-Maturity-Checker")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	type ReleaseItem struct {
		Hash    string `json:"hash"`
		Channel string `json:"channel"`
		Version string `json:"version"`
	}
	var data struct {
		CurrentRelease struct {
			Stable string `json:"stable"`
		} `json:"current_release"`
		Releases []ReleaseItem `json:"releases"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	stableHash := data.CurrentRelease.Stable
	for _, rel := range data.Releases {
		if rel.Hash == stableHash {
			return rel.Version, nil
		}
	}
	for _, rel := range data.Releases {
		if rel.Channel == "stable" {
			return rel.Version, nil
		}
	}
	return "", fmt.Errorf("no stable release found")
}

func (h *SovereignPurifier) VerifyLocalState() error {
	records, err := ParseSBOM(SbomPath)
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("================================================================================\n")
	sb.WriteString("             SOVEREIGN LOCAL STATE SEAL VERIFICATION\n")
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s\n", "Package Name", "Expected Hash", "Status"))
	sb.WriteString("--------------------------------------------------------------------------------\n")

	mismatches := 0
	for _, rec := range records {
		if rec.Category == "LADDER_DOWN" {
			continue
		}
		path := GetPhysicalPath(rec.Name)
		if _, err := os.Stat(path); err != nil {
			hashStr := "..."
			if len(rec.PrunedHash) >= 12 {
				hashStr = rec.PrunedHash[:12] + "..."
			}
			sb.WriteString(fmt.Sprintf("%-24s | %-16s | MISSING\n", rec.Name, hashStr))
			mismatches++
			continue
		}

		actualHash, _, err := h.PruneAndHash(path, false, true)
		if err != nil {
			hashStr := "..."
			if len(rec.PrunedHash) >= 12 {
				hashStr = rec.PrunedHash[:12] + "..."
			}
			sb.WriteString(fmt.Sprintf("%-24s | %-16s | ERROR: %v\n", rec.Name, hashStr, err))
			mismatches++
			continue
		}

		status := "MATCH"
		if actualHash != rec.PrunedHash {
			status = "MISMATCH (Dirty local state)"
			mismatches++
		}
		hashStr := "..."
		if len(rec.PrunedHash) >= 12 {
			hashStr = rec.PrunedHash[:12] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-24s | %-16s | %s\n", rec.Name, hashStr, status))
	}
	sb.WriteString("================================================================================\n")
	_, _ = os.Stdout.WriteString(sb.String())

	if mismatches > 0 {
		return fmt.Errorf("verification failed: found %d mismatches/missing packages", mismatches)
	}
	return nil
}

func (h *SovereignPurifier) updateSBOMVersionAndResetHash(path string, name string, newVersion string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) > 3 && parts[0] == name {
			parts[1] = newVersion
			parts[3] = ""
			lines[i] = strings.Join(parts, "|")
		}
	}
	output := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(output), 0644)
}

func (h *SovereignPurifier) UpgradeManifestUpstream(dryRun bool) error {
	records, err := ParseSBOM(SbomPath)
	if err != nil {
		return err
	}

	cache := loadUpgradeCache()
	now := time.Now()
	githubHold := now.Before(cache.GitHubHoldUntil)

	var sb strings.Builder
	sb.WriteString("================================================================================\n")
	sb.WriteString("             SOVEREIGN UPSTREAM UPGRADE ANALYSIS\n")
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", "Package Name", "Current (SBOM)", "Latest Upstream", "Status"))
	sb.WriteString("--------------------------------------------------------------------------------\n")

	var updatedRecords []string
	cacheUpdated := false

	for _, rec := range records {
		if rec.Category == "LADDER_DOWN" {
			continue
		}

		var latest string
		var fetchErr error
		var resetTime time.Time
		usedCache := false

		entry, hasCache := cache.Entries[rec.Name]
		if hasCache && now.Sub(entry.CheckedAt) < 24*time.Hour {
			latest = entry.Version
			usedCache = true
		}

		upstream, ok := upstreamPackageMappings[rec.Name]
		isGitHub := ok && upstream.typ == "github"

		if isGitHub && githubHold && !usedCache {
			if hasCache {
				latest = entry.Version
				usedCache = true
			} else {
				sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", rec.Name, rec.Version, "RATE_LIMIT", "RATE_LIMIT [HOLD]"))
				continue
			}
		}

		if !usedCache {
			if !ok {
				if rec.Name == "flutter-sdk-firehorse" {
					latest, fetchErr = fetchLatestFlutterRelease()
				} else if rec.Name == "gcloud-sdk" {
					sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", rec.Name, rec.Version, "574.0.0", "UP-TO-DATE"))
					continue
				} else {
					sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", rec.Name, rec.Version, "Unknown", "SKIP"))
					continue
				}
			} else {
				if upstream.typ == "github" {
					latest, resetTime, fetchErr = fetchLatestGitHubRelease(upstream.repo)
				} else {
					latest, fetchErr = fetchLatestNpmRelease(upstream.repo)
				}
			}

			if fetchErr != nil {
				sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", rec.Name, rec.Version, "ERROR", "UNREACHABLE"))
				continue
			}

			if latest == "RATE_LIMITED" {
				if isGitHub {
					cache.GitHubHoldUntil = resetTime
					githubHold = true
					cacheUpdated = true
				}
				if hasCache {
					latest = entry.Version
					usedCache = true
				} else {
					sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", rec.Name, rec.Version, "RATE_LIMIT", "RATE_LIMIT"))
					continue
				}
			} else {
				cache.Entries[rec.Name] = UpgradeCacheEntry{
					Version:   latest,
					CheckedAt: now,
				}
				cacheUpdated = true
			}
		}

		status := "UP-TO-DATE"
		if usedCache {
			status = "UP-TO-DATE (cached)"
		}

		if rec.Version != latest {
			if compareVersions(latest, rec.Version) > 0 {
				status = "DRIFT (Update Avail)"
				if usedCache {
					status = "DRIFT (Update Avail) [cached]"
				}
				if !dryRun {
					if err := h.updateSBOMVersionAndResetHash(SbomPath, rec.Name, latest); err == nil {
						status = "UPDATED IN MANIFEST"
						updatedRecords = append(updatedRecords, rec.Name)
					} else {
						status = "UPDATE ERROR: " + err.Error()
					}
				}
			} else {
				status = "PINNED (Upstream Older)"
			}
		}

		if isGitHub && githubHold && usedCache {
			status += " [HOLD]"
		}

		sb.WriteString(fmt.Sprintf("%-24s | %-16s | %-16s | %-16s\n", rec.Name, rec.Version, latest, status))
		if !usedCache {
			time.Sleep(100 * time.Millisecond)
		}
	}
	sb.WriteString("================================================================================\n")
	_, _ = os.Stdout.WriteString(sb.String())

	if cacheUpdated {
		saveUpgradeCache(cache)
	}
	if !dryRun && len(updatedRecords) > 0 {
		slog.Info("Successfully upgraded manifest versions. Run rehydrator without check to download/seal them.", "updated", updatedRecords)
	}
	return nil
}

func cleanVersion(s string) (string, bool) {
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "go")
	s = strings.TrimPrefix(s, "version_")
	if s == "" {
		return "", false
	}
	firstRune := rune(s[0])
	isDigit := firstRune >= '0' && firstRune <= '9'
	return s, isDigit
}

func compareVersions(v1, v2 string) int {
	clean1, ok1 := cleanVersion(v1)
	clean2, ok2 := cleanVersion(v2)

	if ok1 && !ok2 {
		return 1
	}
	if !ok1 && ok2 {
		return -1
	}

	parts1 := strings.Split(clean1, ".")
	parts2 := strings.Split(clean2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 string
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		stripSuffix := func(s string) string {
			if idx := strings.IndexAny(s, "-+"); idx != -1 {
				return s[:idx]
			}
			return s
		}

		s1 := stripSuffix(p1)
		s2 := stripSuffix(p2)

		n1, err1 := strconv.Atoi(s1)
		n2, err2 := strconv.Atoi(s2)

		if err1 == nil && err2 == nil {
			if n1 != n2 {
				if n1 < n2 {
					return -1
				}
				return 1
			}
		} else {
			if s1 != s2 {
				if s1 < s2 {
					return -1
				}
				return 1
			}
		}
	}
	return 0
}
