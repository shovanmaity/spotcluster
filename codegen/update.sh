#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

(
  cd vendor/k8s.io/code-generator/ 
  go install ./cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen,conversion-gen,defaulter-gen}
)

function codegen::join() { local IFS="$1"; shift; echo "$*"; }

module_name="github.com/shovanmaity/spotcluster"

deepcopy_inputs=(
  pkg/apis/spotcluster.io/v1alpha1 \
)

client_subpackage="pkg/client"
client_package="${module_name}/${client_subpackage}"
client_inputs=(
  pkg/apis/spotcluster.io/v1alpha1 \
)

gen-deepcopy() {
  echo "Generating deepcopy methods..." >&2
  prefixed_inputs=( "${deepcopy_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/deepcopy-gen" \
    --input-dirs "$joined" \
    --output-file-base zz_generated.deepcopy \
    --bounding-dirs "${module_name}"
}

gen-clientsets() {
  echo "Generating clientset..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/client-gen" \
    --clientset-name versioned \
    --input-base "" \
    --input "$joined" \
    --output-package "${client_package}"/clientset
}

gen-listers() {
  echo "Generating listers..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/lister-gen" \
    --input-dirs "$joined" \
    --output-package "${client_package}"/listers
}

gen-informers() {
  echo "Generating informers..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/informer-gen" \
    --input-dirs "$joined" \
    --versioned-clientset-package "${client_package}"/clientset/versioned \
    --listers-package "${client_package}"/listers \
    --output-package "${client_package}"/informers
}

gen-deepcopy
gen-clientsets
gen-listers
gen-informers
