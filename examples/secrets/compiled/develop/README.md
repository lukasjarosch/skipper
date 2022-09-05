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
    "gitlab": {
      "common": {
        "base_url": "https://mygitlab.example.com",
        "project_id": 12345
      }
    },
    "target": {
      "skipper": {
        "secrets": {
          "keys": {
            "aes": "thisis32bitlongpassphraseimusing"
          }
        },
        "use": [
          "azure.common",
          "azure.resources",
          "gitlab.common",
          "terraform.identifiers",
          "terraform.common"
        ]
      }
    },
    "terraform": {
      "common": {
        "backend": {
          "address": "https://mygitlab.example.com",
          "aesTest": "Zw4SlFiRrWnTz1kTZO65q7gkEpkqy7YE",
          "multipleSecrets": "ThisIsMySecret---AnotherSecretValueYay",
          "newDriver": "tfAzKMiqzGFf2Rg8agjPw_ie6A6DSCn_",
          "nonExistingSecretWithAlternativeAction": "YYowXUlOmc0vsbEI9twsa1f6FeI9wTLfRtf9XzrChslW9exqfPqXZoLEk3RVlgYG",
          "password": "ThisIsMySecret",
          "state_name": "develop.tfstate"
        },
        "version": "\u003e= 0.14"
      },
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
