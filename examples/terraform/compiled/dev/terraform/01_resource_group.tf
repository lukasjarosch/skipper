// This code is part of the project 'terraform-example'
// Copyright 2022, AcmeCorp International
//
// This code is generated; DO NOT EDIT.

resource "azurerm_resource_group" "example" {
  location = "westeurope"
  name     = "rg-dev-terraform"
  tags = {
    service   = "terraform-example"
  }
}
