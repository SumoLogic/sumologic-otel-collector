# Set PowerShell as the default shell
$powerShellPath = get-command powershell
New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell -Value $powerShellPath.Path -PropertyType String -Force

# Install Chocolatey
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iwr https://community.chocolatey.org/install.ps1 -UseBasicParsing | iex
choco upgrade chocolatey -r

# Install Golang
choco install golang --version 1.21.2 -yr --no-progress

# Install tools necessary to build otel
choco install gnuwin32-coreutils.install -yr --no-progress
choco install make git -yr --no-progress

# Ensure bash is on the PATH
$path = [Environment]::GetEnvironmentVariable("Path", "User")
$path += ';C:\Program Files\Git\bin'
[Environment]::SetEnvironmentVariable("Path", $path, "User")

# Make git stop complaining about permissions in the shared directory
C:\Program` Files\Git\cmd\git.exe config --global --add safe.directory '%(prefix)///vboxsvr/sumologic/'

# Restart SSH daemon to ensure it loads the new environment variables
Restart-Service sshd
