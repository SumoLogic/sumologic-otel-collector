// https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {}

group "default" {
  targets = [
    "standard",
    "standard-fips",
    "ubi",
    "ubi"
  ]
}

target "_common" {
  inherits = ["docker-metadata-action"]
  context = "./"
  dockerfile = "Dockerfile"
}

target "_common-fips" {
  args = {
    COLLECTOR_BIN = "otelcol-sumo-fips"
  }
}

target "_common-ubi" {
  inherits = ["_common"]
  dockerfile = "Dockerfile_ubi"
}

target "_common-windows" {
  inherits = ["_common"]
  dockerfile = "Dockerfile_windows"
}

target "standard" {
  inherits = ["_common"]
  platforms  = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "standard-fips" {
  inherits = ["_common", "_common-fips"]
  platforms  = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "ubi" {
  inherits = ["_common-ubi"]
  platforms  = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "ubi-fips" {
  inherits = ["_common-ubi", "_common-fips"]
  platforms  = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "windows-ltsc2022" {
  inherits = ["_common-windows"]
  args = {
    BASE_IMAGE_TAG = "ltsc2022"
  }
}

target "windows-ltsc2025" {
  inherits = ["_common-windows"]
  args = {
    BASE_IMAGE_TAG = "ltsc2025"
  }
}

target "windows-ltsc2022-fips" {
  inherits = ["_common-windows", "_common-fips"]
  args = {
    BASE_IMAGE_TAG = "ltsc2022"
  }
}

target "windows-ltsc2025-fips" {
  inherits = ["_common-windows", "_common-fips"]
  args = {
    BASE_IMAGE_TAG = "ltsc2025"
  }
}
