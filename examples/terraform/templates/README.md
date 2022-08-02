{{ $inv := .Inventory }}

# Readme for {{ $inv.target.name }}

{{ $inv.gitlab.project_id }}
{{ $inv.target.azure.resource_group}}

something: {{ $inv.gitlab.something }}

{{ foo }}

{{ .Additional }}

