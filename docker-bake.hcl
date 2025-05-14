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
  output = [{
    type="image",
    name="${BASE_TAG}/sumologic-otel-collector-ci-builds",
    push-by-digest=true,
    name-canonical=true,
    push=true
  }]
  tags = [
    #"663229565520.dkr.ecr.us-east-1.amazonaws.com/sumologic/sumologic-otel-collector-ci-builds:${GIT_SHA}"
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:${CONTAINER_VERSION}",
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:${GIT_SHA}",
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:standard-${GIT_SHA}",
    #"${BASE_TAG}:${GIT_SHA}-standard"
  ]
}

target "standard-fips" {
  inherits = ["standard"]
  args = {
    COLLECTOR_BIN = "otelcol-sumo-fips"
  }
  output = [{
   type="image",
   #name="${BASE_TAG}/sumologic-otel-collector-ci-builds:${GIT_SHA}",
   name="${BASE_TAG}/sumologic-otel-collector-ci-builds",
   push-by-digest=true,
   name-canonical=true,
   push=true
  }]
}

target "standard-local" {
  inherits = ["standard"]
  output = ["type=docker"]
}

target "standard-local-fips" {
  inherits = ["standard-fips"]
  output = ["type=docker"]
}

target "standard-all-nofips" {
  inherits = ["standard"]
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "standard-all-fips" {
  inherits = ["standard-fips"]
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "standard-all" {
  name = "standard-all-${tgt}"
  inherits = [tgt]
  matrix = {
    tgt = [
      "standard-all-nofips",
      "standard-all-fips"
    ]
  }
}

target "push-standard-ecr" {

}


  #tags = [
    #"663229565520.dkr.ecr.us-east-1.amazonaws.com/sumologic/sumologic-otel-collector-ci-builds:${GIT_SHA}"
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:${GIT_SHA}",
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:${CONTAINER_VERSION}",
    #"${BASE_TAG}/sumologic-otel-collector-ci-builds:latest",
  #]
#}
