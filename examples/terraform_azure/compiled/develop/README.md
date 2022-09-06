# Skipper Terraform Azure
> Target: **develop**
> Subscription: **45e4fca6-f05b-4354-951a-3ea194d2da85**

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
      "common": {
        "absolute_variable": "45e4fca6-f05b-4354-951a-3ea194d2da85",
        "local_variable": "45e4fca6-f05b-4354-951a-3ea194d2da85",
        "subscription_id": "45e4fca6-f05b-4354-951a-3ea194d2da85"
      },
      "resources": {
        "location": "westeurope",
        "resource_group": {
          "name": "rg-develop-terraform-example-westeurope"
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
              "name": "virtual_machines"
            }
          }
        }
      }
    },
    "target": {
      "skipper": {
        "use": [
          "azure.common",
          "azure.resources",
          "terraform.identifiers"
        ]
      }
    },
    "terraform": {
      "identifiers": {
        "resource_group": "changed_identifier",
        "subnets": {
          "virtual_machines": "vms"
        },
        "vnet": "vnet"
      }
    }
  },
  "TargetName": "develop"
}
```
