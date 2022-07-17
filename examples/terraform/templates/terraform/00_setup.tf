{{ $inv := .Inventory -}}
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
    address        = "{{ $inv.gitlab.base_url }}/api/v4/projects/{{ $inv.gitlab.project_id }}/terraform/state/{{ $inv.target.terraform.state_name }}"
    lock_address   = "{{ $inv.gitlab.base_url }}/api/v4/projects/{{ $inv.gitlab.project_id }}/terraform/state/{{ $inv.target.terraform.state_name }}/lock"
    unlock_address = "{{ $inv.gitlab.base_url }}/api/v4/projects/{{ $inv.gitlab.project_id }}/terraform/state/{{ $inv.target.terraform.state_name }}/lock"
    lock_method    = "POST"
    unlock_method  = "DELETE"
    retry_wait_min = 5
    retry_wait_max = 15
    username       = "terraform"
  }
}

provider "azurerm" {
  features {}
  subscription_id = "{{ $inv.target.azure.subscription_id }}"
}
