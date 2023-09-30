az keyvault delete \
  --subscription d62fd2d4-358f-4ddc-9373-7ac8a307f75b \
  --resource-group skipper-keyvault-example \
  --name skipperkeyvaultexample \

az keyvault purge \
  --subscription d62fd2d4-358f-4ddc-9373-7ac8a307f75b \
  --name skipperkeyvaultexample \

az group delete \
  --name skipper-keyvault-example \
  --subscription d62fd2d4-358f-4ddc-9373-7ac8a307f75b \
  --yes 
