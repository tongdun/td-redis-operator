#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
# defines absolute root path
ROOT_PATH=${ROOT_PATH:-$(cd ${SCRIPT_ROOT}; pwd -P)}
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo k8s.io/code-generator)}
CONTROLLER_GEN_PKG=${CONTROLLER_GEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/sigs.k8s.io/controller-tools 2>/dev/null || echo sigs.k8s.io/controller-tools)}
GO_HEADER_FILE=${ROOT_PATH}/hack/boilerplate.go.txt

OUTPUT_DIR=${ROOT_PATH}/_output
mkdir ${OUTPUT_DIR}
CRD_DIR=${ROOT_PATH}/crd

# Register function to be called on EXIT to remove generated binary.
function cleanup {
  rm -rf "${OUTPUT_DIR:-}"
}
trap cleanup EXIT

#EXT_APIS_PKG=${ROOT_PATH#"$GOPATH/src/"}/pkg/apis
EXT_APIS_PKG=td-redis-operator/pkg/apis
#OUTPUT_PKG=${ROOT_PATH#"$GOPATH/src/"}/pkg/client
OUTPUT_PKG=td-redis-operator/pkg/client
# apps:v1,v2 othergroup:v1alpha1,v1alpha2
GROUP_VERSIONS="cache:v1alpha1"

function join() { local IFS="$1"; shift; echo "$*"; }

EXT_APIS=()

for GVs in ${GROUP_VERSIONS}; do
  IFS=: read G Vs <<<"${GVs}"
  # enumerate versions
  for V in ${Vs//,/ }; do
    EXT_APIS+=("${EXT_APIS_PKG}/${G}/${V}")
  done
done


echo "Building controller-gen"
CONTROLLER_GEN="${OUTPUT_DIR}/controller-gen"
go build -v -o "${CONTROLLER_GEN}" ${CONTROLLER_GEN_PKG}/cmd/controller-gen

echo "Generating crd at ${CRD_DIR}"

${CONTROLLER_GEN} \
    crd:crdVersions=v1 \
    paths=${ROOT_PATH}/pkg/apis/... \
    output:crd:dir=${CRD_DIR}



echo "Building register-gen"
REGISTER_GEN="${OUTPUT_DIR}/register-gen"
go build -v -o "${REGISTER_GEN}" ${CODEGEN_PKG}/cmd/register-gen

echo "Generating register func for ${GROUP_VERSIONS}"
${REGISTER_GEN} --input-dirs $(join , "${EXT_APIS[@]}") -o ${ROOT_PATH}/..  --go-header-file ${GO_HEADER_FILE}



echo "Building deepcopy-gen"
DEEPCOPY_GEN="${OUTPUT_DIR}/deepcopy-gen"
go build -v -o "${DEEPCOPY_GEN}" ${CODEGEN_PKG}/cmd/deepcopy-gen

echo "Generating deepcopy funcs for ${GROUP_VERSIONS}"
${DEEPCOPY_GEN} --input-dirs $(join , "${EXT_APIS[@]}") -O zz_generated.deepcopy -o ${ROOT_PATH}/.. --bounding-dirs ${EXT_APIS_PKG} --go-header-file ${GO_HEADER_FILE}



echo "Building defaulter-gen"
DEFAULTER_GEN="${OUTPUT_DIR}/defaulter-gen"
go build -v -o "${DEFAULTER_GEN}" ${CODEGEN_PKG}/cmd/defaulter-gen

echo "Generating defaulters for ${GROUP_VERSIONS}"
${DEFAULTER_GEN}  --input-dirs $(join , "${EXT_APIS[@]}") -O zz_generated.defaults -o ${ROOT_PATH}/.. --go-header-file ${GO_HEADER_FILE}



echo "Building client-gen"
CLIENT_GEN="${OUTPUT_DIR}/client-gen"
go build -v -o "${CLIENT_GEN}" ${CODEGEN_PKG}/cmd/client-gen
echo "Generating clientset for ${GROUP_VERSIONS} at ${OUTPUT_PKG}/clientset"
echo $(join , "${EXT_APIS[@]}")
echo ${ROOT_PATH}
echo ${OUTPUT_PKG}
${CLIENT_GEN} --clientset-name clientset --input-base "" --input $(join , "${EXT_APIS[@]}") -o ${ROOT_PATH}/..  --output-package ${OUTPUT_PKG} --go-header-file ${GO_HEADER_FILE}


echo "Building lister-gen"
LISTER_GEN="${OUTPUT_DIR}/lister-gen"
go build -v -o "${LISTER_GEN}" ${CODEGEN_PKG}/cmd/lister-gen

echo "Generating listers for ${GROUP_VERSIONS} at ${OUTPUT_PKG}/listers"
${LISTER_GEN} --input-dirs $(join , "${EXT_APIS[@]}") --output-package ${OUTPUT_PKG}/listers -o ${ROOT_PATH}/..  --go-header-file ${GO_HEADER_FILE}



echo "Building informer-gen"
INFORMER_GEN="${OUTPUT_DIR}/informer-gen"
go build -v -o "${INFORMER_GEN}" ${CODEGEN_PKG}/cmd/informer-gen

echo "Generating informers for ${GROUP_VERSIONS} at ${OUTPUT_PKG}/informers"
${INFORMER_GEN} \
    --input-dirs $(join , "${EXT_APIS[@]}") \
    --versioned-clientset-package ${OUTPUT_PKG}/clientset \
    --single-directory \
    --listers-package ${OUTPUT_PKG}/listers \
    --output-package ${OUTPUT_PKG}/informers \
    --go-header-file ${GO_HEADER_FILE} \
    -o ${ROOT_PATH}/..

