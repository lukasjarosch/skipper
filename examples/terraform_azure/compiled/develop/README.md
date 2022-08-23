# Skipper Terraform Azure
> Target: **develop**

## Created resources in location `westeurope`

- ResourceGroup: `rg-develop-terraform-example-westeurope`

- Virtual Network: `vnet-develop-terraform-example`
  - AddressSpace: `["10.1.0.0/16", "10.2.0.0/16"]`
  - Subnet `virtual_machines`
    - Address Prefixes: `["10.1.1.0/24"]`


## Template context data available for this target
```json
{
  "Inventory": {
    "azure": {
      "resources": {
        "location": "westeurope",
        "resource_group": {
          "name": "rg-develop-terraform-example-westeurope",
          "tf_identifier": "pizza"
        },
        "vnet": {
          "address_space": [
            "10.1.0.0/16",
            "10.2.0.0/16"
          ],
          "name": "vnet-develop-terraform-example",
          "subnets": {
            "virtual_machines": {
              "address_prefixes": [
                "10.1.1.0/24"
              ],
              "name": "virtual_machines",
              "tf_identifier": "virtual_machines"
            }
          },
          "tf_identifier": "example"
        }
      }
    },
    "target": {
      "use": [
        "azure.resources"
      ]
    }
  },
  "TargetName": "develop"
}
```
