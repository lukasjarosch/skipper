{{ $t := .Target -}}
{{ $i := .Inventory -}}
// Copyright (c) 2022, The CloudServices Team, Markant Services International GmbH
// 
// Project: https://gitlab.markant.com/cloud-services/azure/edge-v2/generator
// Version: develop (e21d3f6)
// Timestamp: 20 Jun 22 09:27 CEST
// 
// This code is generated; DO NOT EDIT.

resource "azurerm_resource_group" "edge" {
  location = "{{ $t.azure.location }}"
  name     = "{{ $t.azure.resource_group }}"
  tags = {
    owner     = "cloud-services"
    service   = "{{ $i.project.name }}"
  }
}
