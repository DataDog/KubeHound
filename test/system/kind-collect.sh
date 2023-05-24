#!/bin/bash
#
# This script extracts the information from all pods from all namespaces in all clusters in json files.
# It extracts the info from one namespace after another to avoid overloading the API.
# The *_ALLOWLIST variables can be used to only extract data from a subset of clusters/namespaces/resources.
#

#
# ALLOWLIST SETTINGS
#
# patterns matching cluster names we want to collect, customize to your needs
CLUSTERS_ALLOWLIST=(
    kind-kubehound.test.local
)
# patterns matching namespaces we want to collect, customize to your needs
NAMESPACES_ALLOWLIST=(
    *
)
# patterns matching metrics we want to collect, customize to your needs (you don't need to change that to use with pyscanner)
RESOURCES_ALLOWLIST=(
    pods
    roles*
    rolebinding*
)

CLUSTER_RESOURCES=(
    nodes
    clusterroles.rbac.authorization.k8s.io
    clusterrolebindings.rbac.authorization.k8s.io
)

#

#
# CLI FLAGS VARIABLES
#
ALLOWLIST_MODE=1
# fetch only missing resources if 1
UPDATE=0
# how long should the script wait between 2 calls to "kubectl get <resource> -n <namespace> -o json"
DELAY=2
#

#
# INTERNAL VARIABLES
#
# how long should the script wait before starting the extraction
STARTUP_DELAY=5
# temporary file for storing kubectl errors
ERRFILE="/tmp/get-cluster-data-errors.log"
# patterns matching resources we don't want to collect (yet?)
RESOURCES_DENYLIST=(
    *.istio.io
    datadogmetric*
    *.containo.us
)
#

# utility functions
function red() {
    echo -ne "\033[1;31m${1}\033[0m"
}

function green() {
    echo -ne "\033[1;32m${1}\033[0m"
}

# arguments handling helpers
function show_usage() {
    cat << EOF
Usage ${0} [-ua] [-d DELAY] OUTDIR
    -u              Update partial extract, only fetch missing resources
    -d DELAY        Delay between api calls (default: ${DELAY}sec)
    -a              Allowlist mode
    -h              Show this help message
EOF
}

# record start date and parse arguments
start_time="$(date +'%Y-%m-%d %H:%M:%S')"
OPTIND=1
while getopts "d:uah" opt; do
    case "${opt}" in
        d)
            DELAY="${OPTARG}"
            ;;
        u)
            UPDATE=1
            ;;
        a)
            ALLOWLIST_MODE=1
            ;;
        *)
            show_usage
            exit 1
            ;;
    esac
done
shift "$((OPTIND-1))"
if [ $# -ne 1 ]; then
    show_usage
    exit 1
fi
OUTDIR="${1}"

# prepare list of target clusters
ALL_CLUSTERS=$(kubectl config get-contexts -o name)
TARGET_CLUSTERS=()
for cluster in ${ALL_CLUSTERS}; do
    if [[ ${ALLOWLIST_MODE} -eq 1 ]]; then
        for pattern in "${CLUSTERS_ALLOWLIST[@]}"; do
            if [[ ${cluster} == ${pattern} ]]; then
                TARGET_CLUSTERS+=("${cluster}")
                break
            fi
        done
    else
        TARGET_CLUSTERS+=("${cluster}")
    fi
done

# give grace time to stop the process
echo "Output folder: $(green "${OUTDIR}")"
echo "Delay between calls to kubectl: $(green "${DELAY}sec")"
echo "Target clusters:"
for target in "${TARGET_CLUSTERS[@]}"; do
    echo "  - $(green "${target}")"
done
echo "Starting extraction in ${STARTUP_DELAY}sec"
sleep ${STARTUP_DELAY}

# create output folder
if [[ -d "${OUTDIR}" ]] && [ ${UPDATE} -eq 0 ]; then
    echo "Output folder already exists: $(red "${OUTDIR}")"
    exit 1
fi
if [[ ! -d "${OUTDIR}" ]]; then
    echo "Creating output folder: $(green "${OUTDIR}")"
    mkdir -p "${OUTDIR}"
fi

echo ""

# loop over clusters
for cluster in "${TARGET_CLUSTERS[@]}"; do
    # switch to cluster
    echo "Connecting to the $(green "${cluster}") cluster"
    kubectl config use-context "${cluster}"
    if [[ $? -ne 0 ]]; then
        echo "Could not switch to the $(red "${cluster}") cluster, aborting"
        echo "Your kubectl config might be outdated"
        exit 1
    fi

    # handle cluster-level resources
    cluster_dir="${OUTDIR}/${cluster}"
    mkdir "${cluster_dir}"
    echo -n "Extracting cluster resources..."
    for resource in "${CLUSTER_RESOURCES[@]}"; do
        outfile="${cluster_dir}/${resource}.json"
        sleep "${DELAY}"
        echo -n kubectl get "${resource}" -o json > "${outfile}" 2> "${ERRFILE}"
        kubectl get "${resource}" -o json > "${outfile}" 2> "${ERRFILE}"
    done

    # extract resources
    echo -n "Extracting resources list..."
    resources=$(kubectl api-resources -o name)
    res_count="$(echo "${resources}" | wc -l)"
    echo " $(green "${res_count}") found"

    # extract namespaces
    echo -n "Extracting namespaces list..."
    namespaces=$(kubectl get ns -o custom-columns=NAME:.metadata.name --no-headers)
    ns_count="$(echo "${namespaces}" | wc -l)"
    echo " $(green "${ns_count}") found"

    echo ""

    # loop over namespaces
    ns_current=1
    for namespace in ${namespaces}; do
        if [[ ${ALLOWLIST_MODE} -eq 1 ]]; then
            wanted=1
            for pattern in "${NAMESPACES_ALLOWLIST[@]}"; do
                if [[ ${namespace} == ${pattern} ]]; then
                    wanted=1
                    break
                fi
            done
            if [[ ${wanted} -ne 1 ]]; then
                echo "Namespace $(red "${namespace}") ($(green "${ns_current}")/${ns_count}) - does not match allowlist (skipped)"
                ns_current=$((ns_current+1))
                continue
            fi
        fi

        echo "Extracting resources from the $(green "${namespace}") namespace in the $(green "${cluster}") cluster ($(green "${ns_current}")/${ns_count})"
        ns_current=$((ns_current+1))
        res_current=1

        for resource in ${resources}; do
            # check if resource is unwanted
            unwanted=0
            for pattern in "${RESOURCES_DENYLIST[@]}"; do
                if [[ ${resource} == ${pattern} ]]; then
                    echo "  - $(red "${resource}") ($(green "${res_current}")/${res_count}) - matches denylist (skipped)"
                    unwanted=1
                    res_current=$((res_current+1))
                    break
                fi
            done
            [[ ${unwanted} -eq 1 ]] && continue

            # check if resource is wanted (if allowlist mode)
            if [[ ${ALLOWLIST_MODE} -eq 1 ]]; then
                wanted=0
                for pattern in "${RESOURCES_ALLOWLIST[@]}"; do
                    if [[ ${resource} == ${pattern} ]]; then
                        wanted=1
                        break
                    fi
                done
                if [[ ${wanted} -ne 1 ]]; then
                    echo "  - $(red "${resource}") ($(green "${res_current}")/${res_count}) - does not match allowlist (skipped)"
                    res_current=$((res_current+1))
                    continue
                fi
            fi

            # create output folder and generate filename
            ns_dir="${cluster_dir}/${namespace}"
            mkdir -p "${ns_dir}"
            outfile="${ns_dir}/${resource}.json"

            # check if data already exists for the resource
            if [[ -f "${outfile}" ]] && [ ${UPDATE} -eq 1 ]; then
                echo "  - $(green "${resource}") ($(green "${res_current}")/${res_count}) - already exists (skipped)"
                res_current=$((res_current+1))
                continue
            fi

            # extract data
            sleep "${DELAY}"
            kubectl get "${resource}" -n "${namespace}" -o json > "${outfile}" 2> "${ERRFILE}"
            # TODO: handle return code

            # check if access was denied
            grep -q "MethodNotAllowed" "${ERRFILE}"
            if [ $? -eq 0 ]; then
                echo "  - $(red "${resource}") ($(green "${res_current}")/${res_count}) - access denied"
                rm "${outfile}"
                res_current=$((res_current+1))
                echo '{"error":true}' > "${outfile}"
                continue
            fi

            # success status
            echo "  - $(green "${resource}") ($(green "${res_current}")/${res_count})"
            res_current=$((res_current+1))
        done
        echo ""
    done
done

# cleanup
rm -f "${ERRFILE}"

# report start and end time
echo "Started: ${start_time}"
echo "Ended: $(date +'%Y-%m-%d %H:%M:%S')"
