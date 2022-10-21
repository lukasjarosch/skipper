az keyvault delete \
  --subscription 59efa773-ee54-47d6-a95a-eac3fca3bc24 \
  --resource-group skipper-example \
  --name skipperkeyvaultexample \

az keyvault purge \
  --subscription 59efa773-ee54-47d6-a95a-eac3fca3bc24 \
  --name skipperkeyvaultexample \

az group delete \
  --name skipper-example \
  --subscription 59efa773-ee54-47d6-a95a-eac3fca3bc24 \
  --yes 
