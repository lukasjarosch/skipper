# Skipper Terraform Azure
> Target: **{{ .TargetName }}**
> Subscription: **{{ .Inventory.azure.common.subscription_id }}**

## Created resources in location `{{ .Inventory.azure.resources.location }}`

- ResourceGroup: `{{ .Inventory.azure.resources.resource_group.name }}`
{{ $vnet := .Inventory.azure.resources.vnet }}
- Virtual Network: `{{ $vnet.name }}`
  - AddressSpace: `[{{ $vnet.address_space | tfStringArray }}]`
  - Subnet `{{ $vnet.subnets.virtual_machines.name }}`
    - Address Prefixes: `[{{ $vnet.subnets.virtual_machines.address_prefixes | tfStringArray }}]`

## Template context data available for this target
```json
{{ . | toPrettyJson }}
```
