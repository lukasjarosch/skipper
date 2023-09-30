az group create \
  --location westeurope \
  --name skipper-keyvault-example \
  --subscription d62fd2d4-358f-4ddc-9373-7ac8a307f75b

az keyvault create \
  --location westeurope \
  --subscription d62fd2d4-358f-4ddc-9373-7ac8a307f75b \
  --resource-group skipper-keyvault-example \
  --name skipperkeyvaultexample

az keyvault set-policy \
  -n skipperkeyvaultexample \
  --key-permissions backup create decrypt delete encrypt get getrotationpolicy import list purge recover release restore rotate setrotationpolicy sign unwrapKey update verify wrapKey \
  --object-id $(az ad signed-in-user show | jq -r '.id')
