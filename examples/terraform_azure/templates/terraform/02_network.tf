resource "azurerm_virtual_network" "{{ .Inventory.terraform.identifiers.vnet }}" {
  name                = "{{ .Inventory.azure.resources.vnet.name }}"
  location            = "{{ .Inventory.azure.resources.location }}"
  resource_group_name = azurerm_resource_group.{{ .Inventory.terraform.identifiers.resource_group }}.name
  address_space       = [{{ .Inventory.azure.resources.vnet.address_space | tfStringArray }}]
  tags = {
    target: "{{ .TargetName }}"
  }
}

{{- $snet := .Inventory.azure.resources.vnet.subnets.virtual_machines }}

resource "azurerm_subnet" "{{ .Inventory.terraform.identifiers.subnets.virtual_machines }}" {
  name                 = "snet-{{ $snet.name }}"
  resource_group_name = azurerm_resource_group.{{ .Inventory.terraform.identifiers.resource_group }}.name
  virtual_network_name =  azurerm_virtual_network.{{ .Inventory.terraform.identifiers.vnet }}.name
  address_prefixes = [{{ $snet.address_prefixes | tfStringArray }}]
}

