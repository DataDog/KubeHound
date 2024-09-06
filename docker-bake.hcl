variable "GO_VERSION" {
  # default ARG value set in Dockerfile
  default = null
}

# Defines the output folder to override the default behavior.
# See Makefile for details, this is generally only useful for
# the packaging scripts and care should be taken to not break
# them.
variable "DESTDIR" {
  default = ""
}
function "outdir" {
  params = [defaultdir]
  result = DESTDIR != "" ? DESTDIR : "${defaultdir}"
}

# Special target: https://github.com/docker/metadata-action#bake-definition
target "meta-helper" {}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
  }
}

group "default" {
  targets = ["binary"]
}

group "validate" {
  targets = ["lint"]
}

target "lint" {
  inherits = ["_common"]
  target = "lint"
  output = ["type=cacheonly"]
}

target "binary" {
  inherits = ["_common"]
  target = "binary"
  output = [outdir("./bin/build")]
  platforms = ["local"]
}

target "binary-cross" {
  inherits = ["binary"]
  platforms = [
    "darwin/amd64",
    "darwin/arm64",
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "windows/amd64",
    "windows/arm64"
  ]
}

target "release" {
  # Overrinding the branch as this target is only being used in the CI
  args = {
    BUILD_BRANCH = "main"
  }
  inherits = ["binary-cross"]
  target = "release"
  output = [outdir("./bin/release")]
}

target "image-cross" {
  inherits = ["meta-helper", "binary"]
  output = ["type=image"]
}
