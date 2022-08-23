# Skipper Terraform Azure
> Target: **{{ .TargetName }}**

## Created resources in location `{{ .Inventory.azure.resources.location }}`

- ResourceGroup: `{{ .Inventory.azure.resources.resource_group.name }}`
- Virtual Network: `{{ .Inventory.azure.resources.vnet.name }}`
