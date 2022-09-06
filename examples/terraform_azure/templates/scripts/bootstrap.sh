#!/usr/bin/env bash

set -Eeuo pipefail
trap cleanup SIGINT SIGTERM ERR EXIT

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

source ${script_dir}/utils.sh
source ${script_dir}/az.sh

#-------
SUBSCRIPTION_ID="{{ .Inventory.azure.common.subscription_id }}"
RESOURCE_GROUP_NAME="{{ .Inventory.azure.resources.resource_group.name }}"
TERRAFORM_RESOURCE_GROUP_NAME="{{ .Inventory.azure.resources.terraform_resource_group.name }}"
AZURE_LOCATION="{{ .Inventory.azure.resources.location }}"
TERRAFORM_STORAGE_ACCOUNT="{{ .Inventory.azure.resources.terraform_storage_account.name }}"
TERRAFORM_STORAGE_ACCOUNT_CONTAINER="{{ .Inventory.azure.resources.terraform_storage_account.state_container.name }}"
#-------

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-v] 

This script is used to bootstrap the {{ .Inventory.common.project_name }} '{{ .TargetName }}' environment.
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

msg "=> ${GREEN}'{{ .Inventory.common.project_name }}' bootstrap script for the '{{ .TargetName}}' environment${NOFORMAT}"
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
