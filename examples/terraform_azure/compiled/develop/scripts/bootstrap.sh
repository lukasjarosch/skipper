#!/usr/bin/env bash

set -Eeuo pipefail
trap cleanup SIGINT SIGTERM ERR EXIT

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

source ${script_dir}/utils.sh
source ${script_dir}/az.sh

#-------
SUBSCRIPTION_ID="59efa773-ee54-47d6-a95a-eac3fca3bc24"
RESOURCE_GROUP_NAME="rg-develop-terraform_example-westeurope"
TERRAFORM_RESOURCE_GROUP_NAME="rg-develop-terraform_example-terraform-westeurope"
AZURE_LOCATION="westeurope"
TERRAFORM_STORAGE_ACCOUNT="storagedevelopterraform"
TERRAFORM_STORAGE_ACCOUNT_CONTAINER="develop_tfstate"
#-------

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-v] 

This script is used to bootstrap the terraform_example 'develop' environment.
It is meant to be executed only once as it will provision everything required for the terraform configs to work.

Available options:

-h, --help      Print this help and exit
-v, --verbose   Print script debug info
EOF
  exit
}

cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  # script cleanup here
}

parse_params() {
  # default values of variables set from params
  flag=0
  param=''

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    -v | --verbose) set -x ;;
    --no-color) NO_COLOR=1 ;;
    -f | --flag) flag=1 ;; # example flag
    -?*) die "Unknown option: $1" ;;
    *) break ;;
    esac
    shift
  done

  return 0
}

parse_params "$@"
setup_colors

msg "=> ${GREEN}'terraform_example' bootstrap script for the 'develop' environment${NOFORMAT}"
cmdExists "az"
cmdExists "jq"

checkUser
checkSubscription ${SUBSCRIPTION_ID}
msg ""
msg "=> Creating ResourceGroup: $RESOURCE_GROUP_NAME   "
#az group create -g ${RESOURCE_GROUP_NAME} -l ${AZURE_LOCATION} --subscription ${SUBSCRIPTION_ID} > /dev/null
msg "=> Creating ResourceGroup: $RESOURCE_GROUP_NAME...${GREEN}SUCCESS${NOFORMAT}"
msg ""
msg "=> Creating ResourceGroup: $TERRAFORM_RESOURCE_GROUP_NAME"
#az group create -g ${TERRAFORM_RESOURCE_GROUP_NAME} -l ${AZURE_LOCATION} --subscription ${SUBSCRIPTION_ID} > /dev/null
msg "=> Creating ResourceGroup: $TERRAFORM_RESOURCE_GROUP_NAME...${GREEN}SUCCESS${NOFORMAT}"
msg ""
msg "=> Creating StorageAccount: $TERRAFORM_STORAGE_ACCOUNT"
# TODO
echo "Created StorageAccount: ID"
msg "=> Creating StorageAccount: $TERRAFORM_STORAGE_ACCOUNT...${GREEN}SUCCESS${NOFORMAT}"
msg ""
msg "=> Creating StorageAccount container: $TERRAFORM_STORAGE_ACCOUNT_CONTAINER"
# TODO
echo "Created container: ID"
msg "=> Creating StorageAccount container: $TERRAFORM_STORAGE_ACCOUNT_CONTAINER...${GREEN}SUCCESS${NOFORMAT}"
