{{ $inv := .Inventory -}}
// This code is part of the project '{{ $inv.project.name }}'
// {{ $inv.project.file_header }}
//
// This code is generated; DO NOT EDIT.

resource "azurerm_kubernetes_cluster" "example" {
  name                = "{{ $inv.terraform.resources.aks.name }}"
  dns_prefix          = "{{ $inv.terraform.resources.aks.name }}"
  location            = "{{ $inv.terraform.resources.aks.location }}"
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
    name           = "{{ $inv.terraform.resources.aks.node_pool.name }}"
    vm_size        = "{{ $inv.terraform.resources.aks.node_pool.vm_size }}"
    node_count     = "{{ $inv.terraform.resources.aks.node_pool.node_count }}" 
    vnet_subnet_id = azurerm_subnet.aks.id
  }
}
