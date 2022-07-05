{{ $t := .Target -}}
{{ $i := .Inventory -}}
// Copyright (c) 2022, The CloudServices Team, Markant Services International GmbH
//
// Project: https://gitlab.markant.com/cloud-services/azure/edge-v2/generator
// Version: develop (e21d3f6)
// Timestamp: 20 Jun 22 09:27 CEST
//
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
    address        = "https://gitlab.markant.com/api/v4/projects/{{ $i.gitlab.project_id }}/terraform/state/{{ $t.terraform.state_name }}"
    lock_address   = "https://gitlab.markant.com/api/v4/projects/{{ $i.gitlab.project_id }}/terraform/state/{{ $t.terraform.state_name }}/lock"
    unlock_address = "https://gitlab.markant.com/api/v4/projects/{{ $i.gitlab.project_id }}/terraform/state/{{ $t.terraform.state_name }}/lock"
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
