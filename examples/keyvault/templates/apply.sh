az group create \
  --location {{ .Inventory.keyvault.location }} \
  --name {{ .Inventory.keyvault.resource_group }} \
  --subscription {{ .Inventory.target.azure.common.subscription_id }}

az keyvault create \
  --location {{ .Inventory.keyvault.location }} \
  --subscription {{ .Inventory.target.azure.common.subscription_id }} \
  --resource-group {{ .Inventory.keyvault.resource_group }} \
  --name {{ .Inventory.keyvault.name }}
