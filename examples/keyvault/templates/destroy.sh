az keyvault delete \
  --subscription {{ .Inventory.target.azure.common.subscription_id }} \
  --resource-group {{ .Inventory.keyvault.resource_group }} \
  --name {{ .Inventory.keyvault.name }} \

az keyvault purge \
  --subscription {{ .Inventory.target.azure.common.subscription_id }} \
  --name {{ .Inventory.keyvault.name }} \

az group delete \
  --name {{ .Inventory.keyvault.resource_group }} \
  --subscription {{ .Inventory.target.azure.common.subscription_id }} \
  --yes 
