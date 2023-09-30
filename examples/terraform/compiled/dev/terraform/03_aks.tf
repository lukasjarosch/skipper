// This code is part of the project 'terraform-example'
// Copyright 2023, AcmeCorp International
//
// This code is generated; DO NOT EDIT.

resource "azurerm_kubernetes_cluster" "example" {
  name                = "aks-dev-terraform-example"
  dns_prefix          = "aks-dev-terraform-example"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.example.name 

  identity {
    type = "SystemAssigned"
  }

  role_based_access_control {
    enabled = true
  }

  network_profile {
    network_plugin = "azure"
    network_policy = "calico"
    outbound_type  = "loadBalancer"
  }

  default_node_pool {
    name           = "pizzahut"
    vm_size        = "Standard_D2as_v4"
    node_count     = "1" 
    vnet_subnet_id = azurerm_subnet.aks.id
  }
}
