{{ $inv := .Inventory -}}

// This code is part of the project '{{ $inv.project.name }}'
// {{ $inv.project.file_header }}
//
// This code is generated; DO NOT EDIT.

resource "azurerm_resource_group" "example" {
  location = "{{ $inv.terraform.resources.resource_group.location }}"
  name     = "{{ $inv.terraform.resources.resource_group.name }}"
  tags = {
    service   = "{{ $inv.project.name }}"
  }
}
