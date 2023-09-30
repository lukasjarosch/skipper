// This code is part of the project 'terraform-example'
// Copyright 2023, AcmeCorp International
//
// This code is generated; DO NOT EDIT.

resource "azurerm_virtual_network" "edge" {
  name                = "vnet-dev-terraform-example"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.example.name
  address_space       = ["10.1.0.0.0/16"]
  tags = {
    service   = "terraform-example"
    environment = "dev"
  }
}

resource "azurerm_subnet" "aks" {
  name                 = "snet-dev-aks"
  virtual_network_name = azurerm_virtual_network.edge.name
  resource_group_name  = azurerm_resource_group.edge.name
  address_prefixes     = ["10.1.1.0/24"]
  service_endpoints    = ["Microsoft.KeyVault"]
}

