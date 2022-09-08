#!/usr/bin/env bash

set -Eeuo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

source ${script_dir}/utils.sh

function checkUser() {
  if ! az account show | jq -e '.user | select(.type == "user")' > /dev/null ; then
    die "you are not logged in as a user in the \`az\` cli.. exiting.."
  fi
  msg "${YELLOW}Make sure you are logged in with the correct account${NOFORMAT}"
  echo "currently logged in user: $(az account show  | jq -r .user.name)"
  msg "(ctrl+c (x2) to cancel, enter to continue)"
  read -r
}

function checkSubscription() {
  msg "${YELLOW}Checking whether the correct subscription is active${NOFORMAT}"
  echo "Current Subscription: $(az account show  | jq -re '.id')"
  echo "Target Subscription: $1"

  if ! az account show > /dev/null | jq -e --arg SUBSCRIPTION_ID "$1" 'select(.id == $SUBSCRIPTION_ID)' ; then
    msg "${RED}Wrong subscription active, attemting to switch${NOFORMAT}"
    if ! $(az account set --subscription $1 2> /dev/null); then
      die "Unable to switch to target subscription. Make sure that you have the correct rights."
    fi
  fi
  msg "${GREEN}Correct subscription active${NOFORMAT}"
}

