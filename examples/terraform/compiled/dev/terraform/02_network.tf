// This code is generated; DO NOT EDIT.

resource "azurerm_virtual_network" "edge" {
  name                = "vnet-dev-nginx-https-westeurope"
  location            = azurerm_resource_group.edge.location
  resource_group_name = azurerm_resource_group.edge.name
  address_space       = ["10.1.0.0.0/16"]
  tags = {
    service   = "edge-v2"
  }
}

resource "azurerm_subnet" "appgw" {
  name                 = "snet-dev-nginx-https-appgw-westeurope"
  virtual_network_name = azurerm_virtual_network.edge.name
  resource_group_name  = azurerm_resource_group.edge.name
  address_prefixes     = ["10.1.1.0/24"]
  service_endpoints    = ["Microsoft.KeyVault"]
}

