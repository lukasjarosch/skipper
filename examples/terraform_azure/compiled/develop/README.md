# Skipper Terraform Azure
> Target: **develop**
> Subscription: **59efa773-ee54-47d6-a95a-eac3fca3bc24**

## Created resources in location `westeurope`

- ResourceGroup: `rg-develop-terraform_example-westeurope`

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
        "absolute_variable": "59efa773-ee54-47d6-a95a-eac3fca3bc24",
        "local_variable": "59efa773-ee54-47d6-a95a-eac3fca3bc24",
        "subscription_id": "59efa773-ee54-47d6-a95a-eac3fca3bc24"
      },
      "resources": {
        "location": "westeurope",
        "resource_group": {
          "name": "rg-develop-terraform_example-westeurope"
        },
        "terraform_resource_group": {
          "name": "rg-develop-terraform_example-terraform-westeurope"
        },
        "terraform_storage_account": {
          "name": "storagedevelopterraform",
          "state_container": {
            "name": "develop_tfstate"
          }
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
    "common": {
      "project_name": "terraform_example"
    },
    "components": {
      "bootstrap": {
        "skipper": {
          "components": [
            {
              "input_paths": [
                "scripts/bootstrap.sh",
                "scripts/utils.sh",
                "scripts/az.sh"
              ],
              "output_path": "1_bootstrap"
            }
          ]
        }
      },
      "documentation": {
        "skipper": {
          "components": [
            {
              "input_paths": [
                "markdown/docs.md"
              ],
              "output_path": "documentation"
            }
          ]
        }
      },
      "scripts": {
        "skipper": {
          "components": [
            {
              "input_paths": [
                "scripts/bootstrap.sh",
                "scripts/utils.sh",
                "scripts/az.sh"
              ],
              "output_path": "scripts"
            }
          ]
        }
      },
      "terraform": {
        "skipper": {
          "components": [
            {
              "input_paths": [
                "terraform/01_resource_group.tf",
                "terraform/02_network.tf"
              ],
              "output_path": "2_terraform"
            }
          ]
        }
      }
    },
    "target": {
      "skipper": {
        "components": [
          {
            "input_paths": [
              "markdown/README.md"
            ],
            "output_path": "/"
          }
        ],
        "use": [
          "common",
          "azure.common",
          "azure.resources",
          "terraform.identifiers",
          "components.*"
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
