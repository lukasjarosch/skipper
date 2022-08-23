resource "azurerm_resource_group" "changed_identifier" {
  location = "westeurope"
  name     = "rg-develop-terraform-example-westeurope"

  tags = {
    target: "develop"
  }
}
