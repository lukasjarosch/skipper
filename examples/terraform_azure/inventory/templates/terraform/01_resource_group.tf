resource "azurerm_resource_group" "{{ .Inventory.azure.resources.resource_group.identifier }}" {
  location = "{{ .Inventory.azure.resources.location }}"
  name     = "{{ .Inventory.azure.resources.resource_group.name }}"

  tags = {
    target: "{{ .TargetName }}"
  }
}
