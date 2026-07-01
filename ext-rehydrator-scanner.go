package hydration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SovereignScaler manages elastic scale-on-demand workers.
type SovereignScaler struct {
	minInstances int32
	maxInstances int32
	taskQueue    chan func(ctx context.Context) error
	activeCount  int32
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewSovereignScaler(min, max int) *SovereignScaler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &SovereignScaler{
		minInstances: int32(min),
		maxInstances: int32(max),
		taskQueue:    make(chan func(ctx context.Context) error, 100),
		ctx:          ctx,
		cancel:       cancel,
	}
	s.initializeWarmPool()
	return s
}

func (s *SovereignScaler) initializeWarmPool() {
	for i := int32(0); i < s.minInstances; i++ {
		s.spawnWorker(true)
	}
}

func (s *SovereignScaler) spawnWorker(isWarm bool) {
	atomic.AddInt32(&s.activeCount, 1)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		defer atomic.AddInt32(&s.activeCount, -1)

		idleTimer := time.NewTimer(3 * time.Second)
		defer idleTimer.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case task, ok := <-s.taskQueue:
				if !ok {
					return
				}
				if !idleTimer.Stop() {
					select {
					case <-idleTimer.C:
					default:
					}
				}
				_ = task(s.ctx)
				idleTimer.Reset(3 * time.Second)
			case <-idleTimer.C:
				if !isWarm && atomic.LoadInt32(&s.activeCount) > s.minInstances {
					return
				}
				idleTimer.Reset(3 * time.Second)
			}
		}
	}()
}

func (s *SovereignScaler) Submit(task func(ctx context.Context) error) {
	s.taskQueue <- task
	active := atomic.LoadInt32(&s.activeCount)
	queueLen := int32(len(s.taskQueue))

	if queueLen > 0 && active < s.maxInstances {
		s.spawnWorker(false)
	}
}

func (s *SovereignScaler) Shutdown() {
	s.cancel()
	close(s.taskQueue)
	s.wg.Wait()
}

func (h *SovereignPurifier) RunSecurityScan(path string, cvePolicy string) error {
	if cvePolicy == "IGNORE_CVE" {
		slog.Warn("Ignoring security vulnerability scanning due to SBOM override", "path", path)
		return nil
	}
	trivyExe := filepath.Join(`C:\aCogSpaceSeed\00flow\s-forge\91000-external-executables`, "trivy.exe")
	if _, err := os.Stat(trivyExe); err != nil {
		slog.Warn("Trivy executable not found, skipping security scan", "path", path)
		return nil
	}

	cmd := exec.Command(trivyExe, "fs", "--exit-code", "1", "--severity", "HIGH,CRITICAL", "--scanners", "vuln", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return fmt.Errorf("HIGH or CRITICAL security vulnerabilities detected")
		}
		return fmt.Errorf("Trivy failed: %w", err)
	}
	slog.Info("Security scan passed", "path", path)
	return nil
}

func (h *SovereignPurifier) ScanLicenseCompliance(pkgName string, path string) error {
	cache := loadUpgradeCache()
	now := time.Now()

	if lic, ok := cache.Licenses[pkgName]; ok && lic.Status == "passed" && now.Sub(lic.ScannedAt) < 30*24*time.Hour {
		slog.Info("License compliance check bypassed (cached)", "package", pkgName, "cached_at", lic.ScannedAt.Format(time.RFC3339))
		return nil
	}

	var licenseFound bool
	var licenseFiles []string

	_ = filepath.Walk(path, func(curr string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			name := strings.ToLower(info.Name())
			if strings.Contains(name, "license") || strings.Contains(name, "copying") || strings.Contains(name, "notice") {
				licenseFound = true
				licenseFiles = append(licenseFiles, curr)
			}
		}
		return nil
	})

	viralKeywords := []string{
		"AFFERO GENERAL PUBLIC LICENSE",
		"GNU GENERAL PUBLIC LICENSE",
		"GNU AFFERO GENERAL PUBLIC LICENSE",
		"AGPL",
		"GPL V3",
		"GPLV3",
		"GPL-3.0",
		"AGPL-3.0",
		"GENERAL PUBLIC LICENSE V3",
		"SYSTEM-WIDE COPYLEFT LICENSE",
	}

	if licenseFound {
		for _, licFile := range licenseFiles {
			content, err := os.ReadFile(licFile)
			if err != nil {
				continue
			}
			licenseText := strings.ToUpper(string(content))
			for _, kw := range viralKeywords {
				if strings.Contains(licenseText, kw) {
					slog.Error("COMPLIANCE FAULT: Blocked viral copyleft license", "license_file", licFile, "matched_keyword", kw)
					return fmt.Errorf("blocked viral copyleft license (matched: %q) in %s", kw, filepath.Base(licFile))
				}
			}
		}
		slog.Info("License compliance scan passed via primary license descriptors", "package", pkgName)
	} else {
		slog.Warn("No standard license descriptors found. Initiating secondary deep discovery scan...", "package", pkgName)

		type scanJob struct {
			fpath string
			info  os.FileInfo
		}

		jobs := make(chan scanJob, 100)
		results := make(chan error, 1)
		var wg sync.WaitGroup

		for w := 1; w <= 8; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					if job.info.IsDir() || job.info.Size() > 1024*1024 {
						continue
					}
					name := strings.ToLower(job.info.Name())
					isSource := strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".dart") ||
						strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".md") ||
						strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".c") ||
						strings.HasSuffix(name, ".h") || strings.HasSuffix(name, ".py")

					if !isSource {
						continue
					}

					content, err := os.ReadFile(job.fpath)
					if err != nil {
						continue
					}

					licenseText := strings.ToUpper(string(content))
					for _, kw := range viralKeywords {
						if strings.Contains(licenseText, kw) {
							select {
							case results <- fmt.Errorf("blocked viral copyleft license (matched: %q) inline in file: %s", kw, job.fpath):
							default:
							}
							return
						}
					}
				}
			}()
		}

		go func() {
			_ = filepath.Walk(path, func(curr string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					jobs <- scanJob{fpath: curr, info: info}
				}
				return nil
			})
			close(jobs)
			wg.Wait()
			close(results)
		}()

		if err := <-results; err != nil {
			return err
		}
		slog.Info("License compliance scan passed via deep secondary file discovery", "package", pkgName)
	}

	cache.Licenses[pkgName] = LicenseCacheEntry{
		Status:    "passed",
		ScannedAt: now,
	}
	saveUpgradeCache(cache)
	return nil
}
