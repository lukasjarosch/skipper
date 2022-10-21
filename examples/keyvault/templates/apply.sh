az group create \
  --location {{ .Inventory.keyvault.location }} \
  --name {{ .Inventory.keyvault.resource_group }} \
  --subscription {{ .Inventory.target.azure.common.subscription_id }}

az keyvault create \
  --location {{ .Inventory.keyvault.location }} \
  --subscription {{ .Inventory.target.azure.common.subscription_id }} \
  --resource-group {{ .Inventory.keyvault.resource_group }} \
  --name {{ .Inventory.keyvault.name }}

az keyvault set-policy \
  -n {{ .Inventory.keyvault.name }} \
  --key-permissions backup create decrypt delete encrypt get getrotationpolicy import list purge recover release restore rotate setrotationpolicy sign unwrapKey update verify wrapKey \
  --object-id $(az ad signed-in-user show | jq -r '.id')
