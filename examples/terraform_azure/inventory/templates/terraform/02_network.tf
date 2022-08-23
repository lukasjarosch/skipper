resource "azurerm_virtual_network" "{{ .Inventory.azure.resources.vnet.tf_identifier }}" {
  name                = "{{ .Inventory.azure.resources.vnet.name }}"
  location            = "{{ .Inventory.azure.resources.location }}"
  resource_group_name = azurerm_resource_group.{{ .Inventory.azure.resources.resource_group.tf_identifier }}.name
  address_space       = [{{ .Inventory.azure.resources.vnet.address_space | tfStringArray }}]
  tags = {
    target: "{{ .TargetName }}"
  }
}

{{- $snet := .Inventory.azure.resources.vnet.subnets.virtual_machines }}

resource "azurerm_subnet" "{{ $snet.tf_identifier}}" {
  name                 = "snet-{{ $snet.name }}"
  resource_group_name = azurerm_resource_group.{{ .Inventory.azure.resources.resource_group.tf_identifier }}.name
  virtual_network_name =  azurerm_virtual_network.{{ .Inventory.azure.resources.vnet.tf_identifier }}.name
  address_prefixes = [{{ $snet.address_prefixes | tfStringArray }}]
}

