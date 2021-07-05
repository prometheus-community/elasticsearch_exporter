# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64
DOCKER_REPO  ?= olivernadj

include Makefile.common

DOCKER_IMAGE_NAME       ?= elasticsearch-light-exporter
