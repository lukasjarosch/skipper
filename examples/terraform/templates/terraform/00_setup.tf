{{ $t := .Target -}}
{{ $i := .Inventory -}}
// This code is generated; DO NOT EDIT.

terraform {
  required_version = ">= 0.14"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.90.0"
    }
  }
  backend "http" {
    address        = "{{ $i.gitlab.base_url }}/api/v4/projects/{{ $i.gitlab.project_id }}/terraform/state/{{ $t.terraform.state_name }}"
    lock_address   = "{{ $i.gitlab.base_url }}/api/v4/projects/{{ $i.gitlab.project_id }}/terraform/state/{{ $t.terraform.state_name }}/lock"
    unlock_address = "{{ $i.gitlab.base_url }}/api/v4/projects/{{ $i.gitlab.project_id }}/terraform/state/{{ $t.terraform.state_name }}/lock"
    lock_method    = "POST"
    unlock_method  = "DELETE"
    retry_wait_min = 5
    retry_wait_max = 15
    username       = "terraform"
  }
}

provider "azurerm" {
  features {}
  subscription_id = "{{ $t.azure.subscription_id }}"
}
