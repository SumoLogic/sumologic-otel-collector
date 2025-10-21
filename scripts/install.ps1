# =============================================================
# DEPRECATION NOTICE (October 2025)
# The install.ps1 script has been DEPRECATED and will be removed in a future release.
# It has moved to the dedicated packaging repository:
#   https://github.com/SumoLogic/sumologic-otel-collector-packaging
# Please fetch the latest script from the packaging repo releases:
#   https://github.com/SumoLogic/sumologic-otel-collector-packaging/releases/latest/download/install.ps1
# Do not rely on this in-repo copy for new installations; it will no longer receive updates.
# =============================================================

using assembly System.Net.Http
using namespace System.Net.Http

param (
    # Version is used to override the version of otelcol-sumo to install.
    [string] $Version,

    # InstallationToken is used to pass a Sumo Logic installation token to
    # this script. The default value is set to the value of the
    # SUMOLOGIC_INSTALLATION_TOKEN environment variable.
    [string] $InstallationToken = $env:SUMOLOGIC_INSTALLATION_TOKEN,

    # Tags is used to specify a list of tags for the collector. Specified via a
    # hash table.
    # e.g. -Tags @{ tag1 = "foo" ; tag2 = "bar" }
    [Hashtable] $Tags,

    # InstallHostMetrics is used to install host metric collection.
    [bool] $InstallHostMetrics,

    # Fips is used to download a fips binary installer.
    [bool] $Fips,

    # Specifies wether or not remote management is enabled
    [bool] $RemotelyManaged,

    # Ephemeral option enabled
    [bool] $Ephemeral,

    # The API URL used to communicate with the SumoLogic backend
    [string] $Api
)

$PackageGithubOrg = "SumoLogic"
$PackageGithubRepo = "sumologic-otel-collector-packaging"

##
# Security tweaks
#
# This script requires TLS v1.2 or newer. Due to some versions of Windows not
# using TLS v1.2 or TLS v1.3 by default we must detect if it is enabled and
# attempt to enable it if not:
#
# 1. Determine if enabled security protocols contain an allowed security
#    protocol. If yes then do nothing.
# 2. Find which security protocols from the list of allowed protocols are
#    supported by the system. If none, return an error.
# 3. Enable the found security protocols.
##

# A list of secure protocols that this script supports. Ordered from
# most preferred to least preferred.
$allowedSecurityProtocols = @(
    "Tls13"
    "Tls12"
)

function Test-UsingAllowedProtocol {
    foreach ($protocol in $allowedSecurityProtocols) {
        $securityProtocol = [Net.ServicePointManager]::SecurityProtocol
        $securityProtocols = $securityProtocol.ToString().Split(",").Trim()
        if ($securityProtocols -contains $protocol) {
            return $true
        }
    }
    return $false
}

function Get-AvailableAllowedSecurityProtocols {
    $availableProtocols = @()

    foreach ($allowedProtocol in $allowedSecurityProtocols) {
        $definedProtocol = [Enum]::GetNames([Net.SecurityProtocolType]) -contains $allowedProtocol

        if ($definedProtocol) {
            $availableProtocols += $allowedProtocol
        }
    }

    return $availableProtocols
}

function Enable-SecurityProtocol {
    param (
        [Parameter(Position = 0, Mandatory = $true)]
        [Net.SecurityProtocolType] $protocol
    )

    [Net.ServicePointManager]::SecurityProtocol += $protocol
}

if (!(Test-UsingAllowedProtocol)) {
    $protocols = $allowedSecurityProtocols -join ", "
    Write-Warning "No allowed security protocols are enabled on this system. Allowed protocols: ${protocols}"
    Write-Warning "Detecting available security protocols..."

    $available = Get-AvailableAllowedSecurityProtocols
    if ($available.Count -eq 0) {
        Write-Error "No allowed security protocols are available on this system"
    }

    $availableStr = $available -join ", "
    Write-Warning "Detected allowed security protocols on this system: ${availableStr}"
    Write-Warning "Enabling security protocols: ${availableStr}"

    foreach ($name in $available) {
        Enable-SecurityProtocol([Net.SecurityProtocolType]$name)
    }
}

##
# Main functions
##

# A list of architectures can be found on Microsoft's website:
# https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-processor
Enum Architectures
{
    x86     = 0
    MIPS    = 1
    Alpha   = 2
    PowerPC = 3
    ia64    = 6
    x64     = 9
    ARM64   = 12
}

function Get-OSName
{
    Write-Host "Detecting OS type..."
    $platform = [System.Environment]::OSVersion.Platform

    switch ($platform)
    {
        "Win32NT" {}

        default {
            Write-Error "Unsupported OS type: ${platform}" -ErrorAction Stop
        }
    }

    return $platform
}

function Get-ArchName {
    Write-Host "Detecting architecture..."

    [int] $archId = (Get-CimInstance Win32_Processor)[0].Architecture

    $isDefinedArch = [enum]::IsDefined(([Architectures]), 12)
    if (!$isDefinedArch) {
        Write-Error "Unknown architecture id:`t${archId}" -ErrorAction Stop
    }

    [string] $archName = ""
    [Architectures] $arch = $archId

    switch ($arch)
    {
        x64     { $archName = "x64" }

        default {
            Write-Error "Unsupported architecture:`t${arch}" -ErrorAction Stop
        }
    }

    return $archName
}

function Get-InstalledApplicationVersion {
    $product = Get-CimInstance Win32_Product | Where-Object {
        $_.Name -eq "OpenTelemetry Collector" -and $_.Vendor -eq "Sumo Logic"
    }

    if ($product -eq $null) {
        return
    }

    $installLocation = $product.InstallLocation
    $binPath = "${installLocation}bin\otelcol-sumo.exe"

    if (!(Test-Path -Path $binPath -PathType Leaf)) {
        Write-Warning "Sumo Logic OpenTelemtry Collector is installed but otelcol-sumo.exe could not be found. Continuing as if it were not installed."
        return
    }

    $version = . $binPath --version | Out-String

    $versionRegex = '(\d)\.(\d+)\.(\d+)(.*(\d+))'
    $Matches = [Regex]::Matches($version, $versionRegex)
    $majorVersion = $Matches[0].Groups[1].Value
    $minorVersion = $Matches[0].Groups[2].Value
    $patchVersion = $Matches[0].Groups[3].Value
    $suffix = $Matches[0].Groups[4].Value
    $buildVersion = $Matches[0].Groups[5].Value

    return "${majorVersion}.${minorVersion}.${patchVersion}-sumo-${buildVersion}"
}

function Get-InstalledPackageVersion {
    $package = Get-Package -name "OpenTelemetry Collector"

    if ($package -eq $null) {
        return
    }

    return $package.Version.Replace("-", ".")
}

function Get-Version {
    param (
        [Parameter(Mandatory, Position=0)]
        [ValidateSet("All", "Latest")]
        [string] $Command,

        [Parameter(Mandatory, Position=1)]
        [HttpClient] $HttpClient
    )

    switch ($Command) {
        All {
            $request = [HttpRequestMessage]::new()
            $request.Method = "GET"
            $request.RequestURI = "https://api.github.com/repos/" + $PackageGithubOrg + "/" + $PackageGithubRepo + "/releases"
            $request.Headers.Add("Accept", "application/vnd.github+json")
            $request.Headers.Add("X-GitHub-Api-Version", "2022-11-28")

            $response = $HttpClient.SendAsync($request).GetAwaiter().GetResult()
            if (!($response.IsSuccessStatusCode)) {
                $statusCode = [int]$response.StatusCode
                $reasonPhrase = $response.StatusCode.ToString()
                $errMsg = "${statusCode} ${reasonPhrase}"

                if ($response.Content -ne $null) {
                    $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
                    $errMsg += ": ${content}"
                }

                Write-Error $errMsg -ErrorAction Stop
            }

            $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
            $releases = $content | ConvertFrom-Json
            $filtered = @()

            foreach ($release in $releases) {
                # Skip draft releases
                if ($release.Draft -eq $true) {
                    Write-Debug "Skipping draft release: ${release.Name}"
                    continue
                }

                # Skip prereleases
                if ($release.Prerelease -eq $true) {
                    Write-Debug "Skipping prerelease: ${release.Name}"
                    continue
                }

                $filtered += $release.Tag_name.TrimStart("v")
            }

            return $filtered
        }

        Latest {
            $request = [HttpRequestMessage]::new()
            $request.Method = "GET"
            $request.RequestURI = "https://github.com/" + $PackageGithubOrg + "/" + $PackageGithubRepo + "/releases/latest"
            $request.Headers.Add("Accept", "application/json")

            $response = $HttpClient.SendAsync($request).GetAwaiter().GetResult()
            if (!($response.IsSuccessStatusCode)) {
                $statusCode = [int]$response.StatusCode
                $reasonPhrase = $response.StatusCode.ToString()
                $errMsg = "${statusCode} ${reasonPhrase}"

                if ($response.Content -ne $null) {
                    $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
                    $errMsg += ": ${content}"
                }

                Write-Error $errMsg -ErrorAction Stop
            }

            $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
            $release = $content | ConvertFrom-Json

            return $release.Tag_name.TrimStart("v")
        }
    }
}

function Get-Changelog {
    param (
        [Parameter(Mandatory, Position=0)]
        [HttpClient] $HttpClient
    )

    $request = [HttpRequestMessage]::new()
    $request.Method = "GET"
    $request.RequestURI = "https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/main/CHANGELOG.md"

    $response = $HttpClient.SendAsync($request).GetAwaiter().GetResult()
    if (!($response.IsSuccessStatusCode)) {
        $statusCode = [int]$response.StatusCode
        $reasonPhrase = $response.StatusCode.ToString()
        $errMsg = "${statusCode} ${reasonPhrase}"

        if ($response.Content -ne $null) {
            $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
            $errMsg += ": ${content}"
        }

        Write-Error $errMsg -ErrorAction Stop
    }

    return $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
}

function Show-BreakingChanges {
    param (
        [Parameter(Mandatory, Position=0)]
        [string[]] $Versions,

        [Parameter(Mandatory, Position=1)]
        [string] $Changelog
    )

    $splitChangelog = $Changelog -Split "\n"
    $breakingVersions = @()

    foreach ($version in $Versions) {
        # Get lines matching the following and join them into a string:
        # * ##
        # * ## Breaking
        # * breaking changes
        $matchingRegex = "^## |^### Breaking|breaking changes"
        $matchingLines = (
            $splitChangelog | Select-String -Pattern $matchingRegex
        ) -Join "`r`n"

        # Find the section for $version and get the content between it and the
        # next version section.
        $isBreakingRegex = "(?s)(?<=## \[v${version}\]).*?(?=\r\n## )"
        $isBreakingChange = (
            $matchingLines | Select-String -Pattern $isBreakingRegex
        ).Matches.Value -ne ""

        if ($isBreakingChange) {
            $breakingVersions += $version
        }
    }

    if ($breakingVersions.Count -gt 0) {
        $versionsStr = $breakingVersions -Join ", v"
        Write-Host "The following versions contain breaking changes: v${versionsStr}! Please make sure to read the linked Changelog file."
    }
}

function Get-BinaryFromUri {
    param (
        [Parameter(Mandatory, Position=0)]
        [string] $Uri,

        [Parameter(Mandatory, Position=1)]
        [string] $Path,

        [Parameter(Mandatory, Position=2)]
        [HttpClient] $HttpClient
    )

    if (Test-Path $Path) {
        Write-Host "${Path} already exists, removing..."
        Remove-Item $Path
    }

    Write-Host "Preparing to download ${Uri}"
    $requestUri = [System.Uri]$Uri
    $optReadHeaders = [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead
    $response = $HttpClient.GetAsync($requestUri, $optReadHeaders).GetAwaiter().GetResult()
    $responseMsg = $response.EnsureSuccessStatusCode()

    $httpStream = $response.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
    $fileStream = [System.IO.FileStream]::new(
        $Path,
        [System.IO.FileMode]::Create,
        [System.IO.FileAccess]::Write
    )

    $copier = $httpStream.CopyToAsync($fileStream)
    Write-Host "Downloading ${requestUri}"
    $copier.Wait()
    $fileStream.Close()
    $httpStream.Close()

    Write-Host "Downloaded ${Path}"
}

##
# Main code
##

try {
    if ($InstallationToken -eq $null -or $InstallationToken -eq "") {
        Write-Error "Installation token has not been provided. Please set the SUMOLOGIC_INSTALLATION_TOKEN environment variable." -ErrorAction Stop
    }

    $osName = Get-OSName
    $archName = Get-ArchName
    Write-Host "Detected OS type:`t${osName}"
    Write-Host "Detected architecture:`t${archName}"

    $handler = New-Object HttpClientHandler
    $handler.AllowAutoRedirect = $true

    $httpClient = New-Object System.Net.Http.HttpClient($handler)
    $userAgentHeader = New-Object System.Net.Http.Headers.ProductInfoHeaderValue("otelcol-sumo-installer", "0.1")
    $httpClient.DefaultRequestHeaders.UserAgent.Add($userAgentHeader)

    # set http client timeout to 30 seconds
    $httpClient.Timeout = New-Object System.TimeSpan(0, 0, 30)

    if ($Fips -eq $true) {
        if ($osName -ne "Win32NT" -or $archName -ne "x64") {
            Write-Error "Error: The FIPS-approved binary is only available for windows/amd64"
        }
    }

    Write-Host "Getting installed version..."
    $installedAppVersion = Get-InstalledApplicationVersion
    $installedAppVersionStr = "none"
    if ($installedAppVersion -ne $null) {
        $installedAppVersionStr = $installedAppVersion
    }
    $installedPackageVersion = Get-InstalledPackageVersion
    $installedPackageVersionStr = "none"
    if ($installedPackageVersion -ne $null) {
        $installedPackageVersionStr = $installedPackageVersion
    }
    Write-Host "Installed app version:`t${installedAppVersionStr}"
    Write-Host "Installed package version:`t${installedPackageVersionStr}"

    # Get versions, but ignore errors as we fallback to other methods later
    Write-Host "Getting versions..."
    $versions = Get-Version -Command All -HttpClient $httpClient

    # Use user's version if set, otherwise get latest version from API (or website)
    if ($Version -eq "") {
        if ($versions.Count -eq 1) {
            $Version = $versions
        } elseif ($versions.Count -gt 1) {
            $Version = $versions[0]
        } else {
            $Version = Get-Version -Command Latest -HttpClient $httpClient
        }
    }

    # tags in the packaging repository have a dash before the build number, the Windows convention is a stop
    $Tag = $Version
    $Version = $Version.Replace("-", ".")

    Write-Host "Package version to install:`t${Version}"

    # Check if otelcol is already in newest version
    if ($installedPackageVersion -eq $Version) {
        Write-Host "OpenTelemetry collector is already in newest (${Version}) version"
    } else {
        # add newline before breaking changes and changelog
        Write-Host ""
        if (($installedVersion -ne "") -And ($versions -ne $null)) {
            # Take versions from installed up to the newest
            $index = $versions.IndexOf($installedVersion)
            if ($index -gt 0) {
                $changelog = Get-Changelog $httpClient
                Show-BreakingChanges $versions[0..($index-1)] $changelog
            }
        }
    }

    Write-Host "Changelog:`t`thttps://github.com/SumoLogic/sumologic-otel-collector/blob/main/CHANGELOG.md"
    # add newline after breaking changes and changelog
    Write-Host ""

    # Add -fips to the msi filename if necessary
    $fipsSuffix = ""
    if ($Fips -eq $true) {
        Write-Host "Getting FIPS-compliant binary"
        $fipsSuffix = "-fips"
    }

    # Download MSI
    $msiLanguage = "en-US"
    $msiFileName = "otelcol-sumo_${Version}_${msiLanguage}.${archName}${fipsSuffix}.msi"
    $msiUri = "https://github.com/" + $PackageGithubOrg + "/" + $PackageGithubRepo + "/releases/download/"
    $msiUri += "v${Tag}/${msiFileName}"
    $msiPath = "${env:TEMP}\${msiFileName}"
    Get-BinaryFromUri $msiUri -Path $msiPath -HttpClient $httpClient

    # Install MSI
    [string[]] $msiProperties = @()
    [string[]] $msiAddLocal = @()
    $msiProperties += "INSTALLATIONTOKEN=${InstallationToken}"
    if ($Tags.Count -gt 0) {
        [string[]] $tagStrs = @()
        $Tags.GetEnumerator().ForEach({
            $tagStrs += "$($_.Key)=$($_.Value)"
        })
        $tagsProperty = $tagStrs -Join ","
        $msiProperties += "TAGS=`"${tagsProperty}`""
    }
    if ($Api.Length -gt 0) {
        $msiProperties += "API=`"${Api}`""
    }
    if ($InstallHostMetrics -eq $true) {
        $msiAddLocal += "HOSTMETRICS"
    }
    if ($RemotelyManaged -eq $true) {
        $msiAddLocal += "REMOTELYMANAGED"
    }
    if ($Ephemeral -eq $true) {
        $msiAddLocal += "EPHEMERAL"
    }
    if ($msiAddLocal.Count -gt 0) {
        $addLocalStr = $msiAddLocal -Join ","
        $msiProperties += "ADDLOCAL=${addLocalStr}"
    }
    msiexec.exe /i "$msiPath" /passive $msiProperties
} catch [HttpRequestException] {
    Write-Error $_.Exception.InnerException.Message
}
