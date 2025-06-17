#################################################################################
# Groups
#################################################################################

group "default" {
  targets = [
    "standard",
    "standard-fips",
    "ubi",
    "ubi"
  ]
}

#################################################################################
# Base targets
#################################################################################

target "_common" {
  context = "./"
  attest = [
    {
      type = "provenance",
      disabled = true,
    },
    {
      type = "sbom",
      disabled = true,
    },
  ]
}

target "_common-fips" {
  args = {
    COLLECTOR_BIN = "otelcol-sumo-fips"
  }
}

target "_common-standard" {
  inherits = ["_common"]
  dockerfile = "./dockerfiles/scratch/Dockerfile"
}

target "_common-ubi" {
  inherits = ["_common"]
  dockerfile = "./dockerfiles/ubi/Dockerfile"
}

target "_common-windows-2022" {
  inherits = ["_common"]
  dockerfile = "./dockerfiles/windows/nanoserver/ltsc2022/Dockerfile"
}

target "_common-windows-2025" {
  inherits = ["_common"]
  dockerfile = "./dockerfiles/windows/nanoserver/ltsc2025/Dockerfile"
}

#################################################################################
# Composite targets
#################################################################################

target "standard" {
  inherits = ["_common-standard"]
}

target "standard-fips" {
  inherits = ["_common-standard", "_common-fips"]
}

target "ubi" {
  inherits = ["_common-ubi"]
}

target "ubi-fips" {
  inherits = ["_common-ubi", "_common-fips"]
}

target "windows-ltsc2022" {
  inherits = ["_common-windows-2022"]
}

target "windows-ltsc2022-fips" {
  inherits = ["_common-windows-2022", "_common-fips"]
}

target "windows-ltsc2025" {
  inherits = ["_common-windows-2025"]
}

target "windows-ltsc2025-fips" {
  inherits = ["_common-windows-2025", "_common-fips"]
}
