{{ $t := .Target -}}
{{ $i := .Inventory -}}
// Copyright (c) 2022, The CloudServices Team, Markant Services International GmbH
// 
// Project: https://gitlab.markant.com/cloud-services/azure/edge-v2/generator
// Version: develop (e21d3f6)
// Timestamp: 20 Jun 22 09:27 CEST
// 
// This code is generated; DO NOT EDIT.

resource "azurerm_virtual_network" "edge" {
  name                = "vnet-{{ $t.name }}-nginx-https-westeurope"
  location            = azurerm_resource_group.edge.location
  resource_group_name = azurerm_resource_group.edge.name
  address_space       = [{{ $t.azure.network.vnet_address_space | tfStringArray }}]
  tags = {
    owner     = "cloud-services"
    service   = "{{ $i.project.name }}"
  }
}

resource "azurerm_subnet" "appgw" {
  name                 = "snet-{{ $t.name }}-nginx-https-appgw-westeurope"
  virtual_network_name = azurerm_virtual_network.edge.name
  resource_group_name  = azurerm_resource_group.edge.name
  address_prefixes     = [{{ $t.azure.network.appgw_snet_address_prefixes | tfStringArray }}]
  service_endpoints    = ["Microsoft.KeyVault"]
}

