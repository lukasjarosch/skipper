az keyvault key create \
  --name "$1" \
  --size 4096 \
  --kty RSA \
  --ops decrypt encrypt \
  --vault-name {{ .Inventory.keyvault.name }}
