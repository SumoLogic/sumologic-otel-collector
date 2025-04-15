variable "BASE_TAG" {
  validation {
    condition = BASE_TAG != ""
    error_message = "The variable 'BASE_TAG' must not be empty. Please set the BASE_TAG environment variable."
  }
}

variable "GIT_SHA" {
  validation {
    condition = GIT_SHA != ""
    error_message = "The variable 'GIT_SHA' must not be empty. Please set the GIT_SHA environment variable."
  }
}

// https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {}

group "default" {
  targets = ["standard-local"]
}

target "standard" {
  inherits = ["docker-metadata-action"]
  context = "./"
  dockerfile = "Dockerfile"
  tags = [
    #"663229565520.dkr.ecr.us-east-1.amazonaws.com/sumologic/sumologic-otel-collector-ci-builds:${GIT_SHA}"
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:${CONTAINER_VERSION}",
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:${GIT_SHA}",
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:standard-${GIT_SHA}",
    #"${BASE_TAG}:${GIT_SHA}-standard"
  ]
}

target "standard-local" {
  inherits = ["standard"]
  output = ["type=docker"]
}

target "standard-all" {
  inherits = ["standard"]
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
  #output = [{
  #  type="image",
  #  name="${REPO}/sumologic-otel-collector-ci-builds:${GIT_SHA}",
  #  push-by-digest=true,
  #  name-canonical=true,
  #  push=false
  #}]
  #tags = [
    #"663229565520.dkr.ecr.us-east-1.amazonaws.com/sumologic/sumologic-otel-collector-ci-builds:${GIT_SHA}"
    #"${REPO}/sumologic-otel-collector-ci-builds:${GIT_SHA}",
    #"${REPO}/sumologic-otel-collector-ci-builds:${CONTAINER_VERSION}",
    #"${REPO}/sumologic-otel-collector-ci-builds:latest",
  #]
}
