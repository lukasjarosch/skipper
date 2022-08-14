{{ $inv := .Inventory -}}
// This code is generated; DO NOT EDIT.

terraform {
  required_version = "{{ $inv.terraform.common.required_version }}"
  required_providers {
  {{- range $provider := $inv.terraform.common.providers }}
    {{ $provider.name }} = {
      source = {{ $provider.source }}
      version = {{ $provider.version }}
    }
  {{ end -}}
  }

  backend "http" {
    address        = "{{ $inv.gitlab.base_url }}/api/v4/projects/{{ $inv.gitlab.project_id }}/terraform/state/{{ $inv.terraform.common.state_name }}"
    lock_address   = "{{ $inv.gitlab.base_url }}/api/v4/projects/{{ $inv.gitlab.project_id }}/terraform/state/{{ $inv.terraform.common.state_name }}/lock"
    unlock_address = "{{ $inv.gitlab.base_url }}/api/v4/projects/{{ $inv.gitlab.project_id }}/terraform/state/{{ $inv.terraform.common.state_name }}/lock"
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
