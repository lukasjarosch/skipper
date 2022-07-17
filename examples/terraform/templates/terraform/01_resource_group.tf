{{ $t := .Target -}}
{{ $i := .Inventory -}}
// This code is generated; DO NOT EDIT.

resource "azurerm_resource_group" "example" {
  location = "{{ $t.azure.location }}"
  name     = "{{ $t.azure.resource_group }}"
  tags = {
    service   = "{{ $i.project.name }}"
  }
}
