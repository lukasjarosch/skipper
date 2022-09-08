resource "azurerm_resource_group" "{{ .Inventory.terraform.identifiers.resource_group }}" {
  location = "{{ .Inventory.azure.resources.location }}"
  name     = "{{ .Inventory.azure.resources.resource_group.name }}"

  tags = {
    target: "{{ .TargetName }}"
  }
}
