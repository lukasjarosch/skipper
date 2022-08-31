terraform {
  required_version = ">= 0.14"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.90.0"
    }
  }
  backend "http" {
    address        = "https://mygitlab.example.com/api/v4/projects/12345/terraform/state/develop.tfstate"
    lock_address   = "https://mygitlab.example.com/api/v4/projects/12345/terraform/state/develop.tfstate/lock"
    unlock_address = "https://mygitlab.example.com/api/v4/projects/12345/terraform/state/develop.tfstate/lock"
    lock_method    = "POST"
    unlock_method  = "DELETE"
    retry_wait_min = 5
    retry_wait_max = 15
    username       = "terraform"
    password       = "ThisIsMySecret"
  }
}

provider "azurerm" {
  features {
  }
  subscription_id = "45e4fca6-f05b-4354-951a-3ea194d2da85"
}

