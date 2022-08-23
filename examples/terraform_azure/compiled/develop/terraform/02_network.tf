resource "azurerm_virtual_network" "example" {
  name                = "vnet-develop-terraform-example"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.pizza.name
  address_space       = ["10.1.0.0/16", "10.2.0.0/16"]
  tags = {
    target: "develop"
  }
}

resource "azurerm_subnet" "virtual_machines" {
  name                 = "snet-virtual_machines"
  resource_group_name  = "rg-develop-terraform-example-westeurope" 
  virtual_network_name =  azurerm_virtual_network.example.name
  address_prefixes = ["10.1.1.0/24"]
}

