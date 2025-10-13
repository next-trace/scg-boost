#!/usr/bin/env bash
# SCG Version Variables - Source of truth for tool versions
# Used by functions.sh and CI workflows

export GO_REQUIRED="${GO_REQUIRED:-1.25.7}"
export GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION:-v2.7.2}"
export GOVULNCHECK_VERSION="${GOVULNCHECK_VERSION:-v1.1.4}"
export GOSEC_VERSION="${GOSEC_VERSION:-v2.22.2}"
export COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-70.0}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
