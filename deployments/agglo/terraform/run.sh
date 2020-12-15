DEPLOYMENT_DIR=${1?Must specify a deployment dir}
OP=${2?Must specify an operation}

BASEDIR=${PWD}

if [[ ! -d ${DEPLOYMENT_DIR} ]]; then
    echo "'${DEPLOYMENT_DIR}' is not a directory"
    exit 1
fi

docker run --rm -v ${BASEDIR}:/workspace \
        -e AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
        -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
        -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
        -w /workspace/${DEPLOYMENT_DIR} \
        -i -t hashicorp/terraform:light ${OP}
        
