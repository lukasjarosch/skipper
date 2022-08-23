resource "azurerm_resource_group" "pizza" {
  location = "westeurope"
  name     = "rg-develop-terraform-example-westeurope"

  tags = {
    target: "develop"
  }
}
