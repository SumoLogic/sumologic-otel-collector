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

target "standard" {
  inherits = ["docker-metadata-action"]
  context = "./"
  dockerfile = "Dockerfile"
}

target "standard-fips" {
  inherits = ["standard"]
  args = {
    COLLECTOR_BIN = "otelcol-sumo-fips"
  }
}

target "ubi" {
  inherits = ["docker-metadata-action"]
  context = "./"
  dockerfile = "Dockerfile_ubi"
}

target "ubi-fips" {
  inherits = ["ubi"]
  args = {
    COLLECTOR_BIN = "otelcol-sumo-fips"
  }
}

target "windows" {
  inherits = ["docker-metadata-action"]
  context = "./"
  dockerfile = "Dockerfile_windows"
}

target "windows-ltsc2019" {
  name = "standard"
  inherits = ["windows"]
  args = {
    BASE_IMAGE_TAG = "ltsc2019"
  }
}

target "windows-ltsc2022" {
  name = "standard"
  inherits = ["windows"]
  args = {
    BASE_IMAGE_TAG = "ltsc2022"
  }
}
