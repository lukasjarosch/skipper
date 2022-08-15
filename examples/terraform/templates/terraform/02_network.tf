{{ $inv := .Inventory -}}
// This code is part of the project '{{ $inv.project.name }}'
// {{ $inv.project.file_header }}
//
// This code is generated; DO NOT EDIT.

resource "azurerm_virtual_network" "edge" {
  name                = "{{ $inv.terraform.resources.virtual_network.name }}"
  location            = "{{ $inv.terraform.resources.virtual_network.location }}"
  resource_group_name = azurerm_resource_group.example.name
  address_space       = [{{ $inv.target.azure.network.vnet_address_space | tfStringArray }}]
  tags = {
    service   = "{{ $inv.project.name }}"
    environment = "{{ $inv.target.name }}"
  }
}

resource "azurerm_subnet" "aks" {
  name                 = "{{ $inv.terraform.resources.aks_subnet.name }}"
  virtual_network_name = azurerm_virtual_network.edge.name
  resource_group_name  = azurerm_resource_group.edge.name
  address_prefixes     = [{{ $inv.target.azure.network.appgw_snet_address_prefixes | tfStringArray }}]
  service_endpoints    = [{{ $inv.terraform.resources.aks_subnet.service_endpoints | tfStringArray }}]
}

