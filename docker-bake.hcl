# docker-bake.hcl
group "default" {
  targets = ["podsh", "agent", "syncdb"]
}

variable "GIT_TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "ghcr.io"
}

variable "REPO" {
  default = "kthcloud/podsh"
}

target "common" {
  cache-from = ["type=gha"]
  cache-to   = ["type=gha,mode=max"]

  args = {
    GIT_TAG = "${GIT_TAG}"
  }

  # platforms = ["linux/amd64", "linux/arm64"]
}

target "podsh" {
  inherits = ["common"]
  dockerfile = "docker/podsh/Dockerfile"

  tags = concat(
    ["${REGISTRY}/${REPO}/podsh:${GIT_TAG}"],
    GIT_TAG == "latest" ? [] : ["${REGISTRY}/${REPO}/podsh:latest"]
  )
}

target "agent" {
  inherits = ["common"]
  dockerfile = "docker/agent/Dockerfile"

  tags = concat(
    ["${REGISTRY}/${REPO}/agent:${GIT_TAG}"],
    GIT_TAG == "latest" ? [] : ["${REGISTRY}/${REPO}/agent:latest"]
  )
}

target "syncdb" {
  inherits = ["common"]
  dockerfile = "docker/syncdb/Dockerfile"
  
  tags = concat(
    ["${REGISTRY}/${REPO}/syncdb:${GIT_TAG}"],
    GIT_TAG == "latest" ? [] : ["${REGISTRY}/${REPO}/syncdb:latest"]
  )
}

