terraform {
  required_version = "{{ .Inventory.terraform.common.version }}"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.90.0"
    }
  }
  backend "http" {
    address        = "{{ .Inventory.terraform.common.backend.address }}/api/v4/projects/{{ .Inventory.gitlab.common.project_id }}/terraform/state/{{ .Inventory.terraform.common.backend.state_name }}"
    lock_address   = "{{ .Inventory.terraform.common.backend.address }}/api/v4/projects/{{ .Inventory.gitlab.common.project_id }}/terraform/state/{{ .Inventory.terraform.common.backend.state_name }}/lock"
    unlock_address = "{{ .Inventory.terraform.common.backend.address }}/api/v4/projects/{{ .Inventory.gitlab.common.project_id }}/terraform/state/{{ .Inventory.terraform.common.backend.state_name }}/lock"
    lock_method    = "POST"
    unlock_method  = "DELETE"
    retry_wait_min = 5
    retry_wait_max = 15
    username       = "terraform"
    password       = "{{ .Inventory.terraform.common.backend.password }}"
  }
}

provider "azurerm" {
  features {
  }
  subscription_id = "{{ .Inventory.azure.common.subscription_id }}"
}

