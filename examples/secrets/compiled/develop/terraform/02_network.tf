resource "azurerm_virtual_network" "vnet" {
  name                = "vnet-develop-terraform-example"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.changed_identifier.name
  address_space       = ["10.1.0.0/16", "10.2.0.0/16"]
  tags = {
    target: "develop"
  }
}

resource "azurerm_subnet" "vms" {
  name                 = "snet-virtual_machines"
  resource_group_name = azurerm_resource_group.changed_identifier.name
  virtual_network_name =  azurerm_virtual_network.vnet.name
  address_prefixes = ["10.1.1.0/24"]
}

