# Unified Hydration Check Script
# Evolved to audit both External Ingestions and Internal Synthesis Targets

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("external", "internal", "all")]
    [string]$Scope = "all"
)

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "            SOVEREIGN UNIFIED HYDRATION CHECK" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$projectRoot = "C:\aCogSpaceSeed"
$hydrationDir = "C:\aCogSpaceSeed\00flow\s-hydration"
$goExe = "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe"

# 1. External Artifacts Check
if ($Scope -eq "external" -or $Scope -eq "all") {
    Write-Host "`n[1/2] Auditing External Artifact Registry (SBOM)..." -ForegroundColor Yellow
    if (Test-Path "$hydrationDir\hydrator.exe") {
        & "$hydrationDir\hydrator.exe" -check
    } else {
        Write-Host "Warning: hydrator.exe not compiled on disk, running dry-run check via compiler..." -ForegroundColor DarkYellow
        & $goExe run "$hydrationDir\cmd\hydrator\main.go" -check
    }
}

# 2. Internal Workspaces Check
if ($Scope -eq "internal" -or $Scope -eq "all") {
    Write-Host "`n[2/2] Auditing Internal Workspaces (DAG & Harnesses)..." -ForegroundColor Yellow
    
    # Establish workspace list from go.work
    if (Test-Path "$hydrationDir\int-rehydrator.exe") {
        # Running check on each workspace harness
        Get-ChildItem -Path "$projectRoot\00flow" -Filter "workspace.harness" -Recurse | ForEach-Object {
            Write-Host "  -> Verifying harness: $($_.FullName)" -ForegroundColor Blue
            & "$hydrationDir\int-rehydrator.exe" -harness $_.FullName -local-only
        }
    } else {
        Write-Host "Warning: int-rehydrator.exe not compiled on disk, running dry-run checks via compiler..." -ForegroundColor DarkYellow
        Get-ChildItem -Path "$projectRoot\00flow" -Filter "workspace.harness" -Recurse | ForEach-Object {
            Write-Host "  -> Verifying harness: $($_.FullName)" -ForegroundColor Blue
            & $goExe run "$hydrationDir\cmd\int-rehydrator\main.go" -harness $_.FullName -local-only
        }
    }
}

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "          UNIFIED HYDRATION CHECK VERIFICATION COMPLETE" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
