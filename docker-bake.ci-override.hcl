#################################################################################
# Functions
#################################################################################

function "cache-from" {
  params = [tgt]
  result = [
    {
      type = "registry"
      ref = "${REPO}:buildcache-${tgt}-${BAKE_LOCAL_PLATFORM}"
    }
  ]
}

function "cache-to" {
  params = [tgt]
  result = [
    {
      type = "registry"
      mode = "max"
      ref = join("-", [
        "${REPO}:buildcache",
        "${tgt}",
        replace(BAKE_LOCAL_PLATFORM, "/", "-")
      ])
    }
  ]
}

#################################################################################
# Base targets & overrides
#################################################################################

# https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {
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

target "_common" {
  inherits = ["docker-metadata-action"]
  output = [
    {
      type = "image"
      name = "${REPO}"
      name-canonical = true
      push = true
      push-by-digest = true
    }
  ]
  args = {
    foo = "${BAKE_LOCAL_PLATFORM}"
  }
}

#################################################################################
# Composite target overrides
#################################################################################

target "standard" {
  cache-from = cache-from("standard")
  cache-to = cache-to("standard")
}

target "standard-fips" {
  cache-from = cache-from("standard-fips")
  cache-to = cache-to("standard-fips")
}

target "ubi" {
  cache-from = cache-from("ubi")
  cache-to = cache-to("ubi")
}

target "ubi-fips" {
  cache-from = cache-from("ubi-fips")
  cache-to = cache-to("ubi-fips")
}

target "windows-ltsc2022" {
  cache-from = cache-from("windows-ltsc2022")
  cache-to = cache-to("windows-ltsc2022")
}

target "windows-ltsc2022-fips" {
  cache-from = cache-from("windows-ltsc2022-fips")
  cache-to = cache-to("windows-ltsc2022-fips")
}

target "windows-ltsc2025" {
  cache-from = cache-from("windows-ltsc2025")
  cache-to = cache-to("windows-ltsc2025")
}

target "windows-ltsc2025-fips" {
  cache-from = cache-from("windows-ltsc2025-fips")
  cache-to = cache-to("windows-ltsc2025-fips")
}
