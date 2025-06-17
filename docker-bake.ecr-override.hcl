#################################################################################
# Variables
#################################################################################

variable "REPO" {
  type = list(string)
  default = join("/", [
    "663229565520.dkr.ecr.us-east-1.amazonaws.com",
    "sumologic/sumologic-otel-collector-ci-builds",
  ])
}
