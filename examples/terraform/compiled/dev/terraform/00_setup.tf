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
    address        = "https://gitlab.example.com/api/v4/projects/1234/terraform/state/dev-state"
    lock_address   = "https://gitlab.example.com/api/v4/projects/1234/terraform/state/dev-state/lock"
    unlock_address = "https://gitlab.example.com/api/v4/projects/1234/terraform/state/dev-state/lock"
    lock_method    = "POST"
    unlock_method  = "DELETE"
    retry_wait_min = 5
    retry_wait_max = 15
    username       = "terraform"
  }
}

provider "azurerm" {
  features {}
  subscription_id = "my-awesome-subscription"
}
