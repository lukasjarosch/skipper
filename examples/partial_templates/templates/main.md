# main template

---

{{ template "no_data" }}

---

{{ . }}

---

{{ template "with_data" context "TargetName" .TargetName "Network" .Inventory.network }}
