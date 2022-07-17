{{ $inv := .Inventory -}}
// This code is generated; DO NOT EDIT.

resource "azurerm_resource_group" "example" {
  location = "{{ $inv.target.azure.location }}"
  name     = "{{ $inv.target.azure.resource_group }}"
  tags = {
    service   = "{{ $inv.project.name }}"
  }
}
