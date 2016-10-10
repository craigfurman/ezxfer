#!/usr/bin/env bash
set -eu
set -o pipefail

ginkgo -r -randomizeSuites -randomizeAllSpecs $@
