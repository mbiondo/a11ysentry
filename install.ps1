# A11ySentry Windows Smart Installer
# Detects Architecture, downloads the latest binary, 
# and registers MCP in Claude, Cursor, etc.

$Owner = "mbiondo"
$Repo = "a11ysentry"
$BinaryName = "a11ysentry.exe"

Write-Host "A11ySentry: Starting smart installation for Windows..." -ForegroundColor Cyan

# 1. Detect Arch
$Arch = "amd64"
if ([IntPtr]::Size -eq 4) { $Arch = "386" }

# 2. Setup Directories
$InstallDir = Join-Path $env:USERPROFILE ".a11ysentry\bin"
if (!(Test-Path $InstallDir)) { 
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# 3. Robust Replacement Strategy
$MoveBinary = {
    param($src, $dest)
    if (Test-Path $dest) {
        Stop-Process -Name "a11ysentry" -Force -ErrorAction SilentlyContinue
        Start-Sleep -Milliseconds 300
        try {
            Move-Item $src -Destination $dest -Force -ErrorAction Stop
        } catch {
            # If locked, rename it (Windows allows renaming running binaries)
            $OldFile = "$dest.old"
            if (Test-Path $OldFile) { Remove-Item $OldFile -Force -ErrorAction SilentlyContinue }
            Rename-Item $dest -NewName "$BinaryName.old" -Force -ErrorAction SilentlyContinue
            Move-Item $src -Destination $dest -Force
        }
    } else {
        Move-Item $src -Destination $dest -Force
    }
}

# 4. Download from GitHub or use local if in DEV MODE
$BinaryFull = Join-Path $InstallDir $BinaryName

# Handle $PSScriptRoot being empty when running via IEX
$LocalBinary = $null
if ($null -ne $PSScriptRoot -and $PSScriptRoot -ne "") {
    $LocalBinary = Join-Path $PSScriptRoot "cmd\a11ysentry\a11ysentry.exe"
}

if ($null -ne $LocalBinary -and (Test-Path $LocalBinary)) {
    & $MoveBinary $LocalBinary $BinaryFull
    Write-Host "DEV MODE: Updated binary from local workspace." -ForegroundColor Yellow
} else {
    Write-Host "Downloading latest release from GitHub..." -ForegroundColor Cyan
    $LatestReleaseUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
    try {
        $ReleaseInfo = Invoke-RestMethod -Uri $LatestReleaseUrl -UseBasicParsing
        $Tag = $ReleaseInfo.tag_name
        $Asset = $ReleaseInfo.assets | Where-Object { $_.name -like "*windows_$Arch.zip" } | Select-Object -First 1
        
        if ($null -eq $Asset) {
            throw "Could not find a valid Windows $Arch asset in the latest release ($Tag)."
        }

        $DownloadUrl = $Asset.browser_download_url
        $TempZip = Join-Path $env:TEMP "a11ysentry.zip"
        
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
        
        # Extract
        Expand-Archive -Path $TempZip -DestinationPath $env:TEMP -Force
        $ExtractedBinary = Join-Path $env:TEMP $BinaryName
        
        if (Test-Path $ExtractedBinary) {
            & $MoveBinary $ExtractedBinary $BinaryFull
        } else {
            $Files = Get-ChildItem -Path $env:TEMP -Filter $BinaryName -Recurse
            if ($Files.Count -gt 0) {
                & $MoveBinary $Files[0].FullName $BinaryFull
            } else {
                throw "Binary not found inside the downloaded archive."
            }
        }
        
        Remove-Item $TempZip -ErrorAction SilentlyContinue
        Write-Host "Successfully installed A11ySentry $Tag." -ForegroundColor Green
    } catch {
        Write-Error "Failed to download or install A11ySentry: $_"
        return
    }
}

# 5. Add to PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    $env:Path += ";$InstallDir"
    Write-Host "Added to User PATH." -ForegroundColor Green
}

# 6. MCP Registration & Skill Setup
Write-Host "Detecting AI Agents and registering MCP..." -ForegroundColor Cyan
$BinaryFull = Join-Path $InstallDir $BinaryName
if (Test-Path $BinaryFull) {
    # Call the binary with the mcp subcommand
    & $BinaryFull mcp --register
    
    # 7. Skill Registration
    $SkillSource = $null
    if ($null -ne $PSScriptRoot -and $PSScriptRoot -ne "") {
        $SkillSource = Join-Path $PSScriptRoot "skills\a11ysentry-mcp"
    }
    
    $AgentSkillsDir = Join-Path $env:USERPROFILE ".gemini\skills\a11ysentry-mcp"
    if ($null -ne $SkillSource -and (Test-Path $SkillSource)) {
        if (!(Test-Path $AgentSkillsDir)) {
            New-Item -ItemType Directory -Path $AgentSkillsDir -Force | Out-Null
        }
        Copy-Item -Path "$SkillSource\*" -Destination $AgentSkillsDir -Recurse -Force
        Write-Host "Registered a11ysentry-mcp skill in Agent Teams." -ForegroundColor Green
    }
} else {
    Write-Warning "Automatic MCP registration skipped (binary not found at $BinaryFull)."
}

Write-Host "A11ySentry installed successfully!" -ForegroundColor Green
Write-Host "Restart your terminal and try running 'a11ysentry --tui'."
