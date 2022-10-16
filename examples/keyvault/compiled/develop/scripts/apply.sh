az group create \
  --location westeurope \
  --name skipper-example \
  --subscription 59efa773-ee54-47d6-a95a-eac3fca3bc24

az keyvault create \
  --location westeurope \
  --subscription 59efa773-ee54-47d6-a95a-eac3fca3bc24 \
  --resource-group skipper-example \
  --name skipperkeyvaultexample

