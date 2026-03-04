# docker-bake.hcl

group "default" {
  targets = ["all"]
}

variable "VERSION" {
  default = "latest"
}

variable "GIT_TAG" {
  default = "unknown"
}

variable "REGISTRY" {
  default = "ghcr.io/"
}

variable "REPO" {
  default = "kthcloud/podsh/"
}

variable "BUILD_VERSIONS" {
  default = "latest,master"
}

target "all" {
  name = "${tgt}-${sanitize(vers)}"
  matrix = {
    tgt = ["podsh", "agent", "syncdb"]
    vers = compact(split(",", "${BUILD_VERSIONS}"))
  }
  inherits = ["${tgt}"]
  args = {
    GIT_TAG = vers
  }
  tags = ["${REGISTRY}${REPO}${tgt}:${vers}"]
}

target "podsh" {
  dockerfile = "docker/podsh/Dockerfile"
  tags = ["${REGISTRY}${REPO}podsh:${VERSION}"] 
  args = {
    GIT_TAG = "${GIT_TAG}"
  }
  cacheFrom = "type=gha"
  cacheTo = "type=gha,mode=max"
  #platforms = ["linux/amd64", "linux/arm64"]
}

target "agent" {
  dockerfile = "docker/agent/Dockerfile"
  tags = ["${REGISTRY}${REPO}agent:${VERSION}"]
  args = {
    GIT_TAG = "${GIT_TAG}"
  }
  cacheFrom = "type=gha"
  cacheTo = "type=gha,mode=max"
  #platforms = ["linux/amd64", "linux/arm64"]
}

target "syncdb" {
  dockerfile = "docker/syncdb/Dockerfile"
  tags = ["${REGISTRY}${REPO}syncdb:${VERSION}"]
  args = {
    GIT_TAG = "${GIT_TAG}"
  }
  cacheFrom = "type=gha"
  cacheTo = "type=gha,mode=max"
  #platforms = ["linux/amd64", "linux/arm64"]
}

