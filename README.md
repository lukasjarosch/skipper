<div align="center">
  <!-- BADGES -->
  <p>
    <a href="https://pkg.go.dev/github.com/lukasjarosch/skipper"><img src="https://pkg.go.dev/badge/github.com/lukasjarosch/skipper.svg" alt="Go Reference"></a>
    <a href="https://goreportcard.com/report/github.com/lukasjarosch/skipper"><img src="https://goreportcard.com/badge/github.com/lukasjarosch/skipper"></a>
  </p>
  <br/>

  <!-- LOGO -->
  <a href="https://github.com/lukasjarosch/skipper">
    <img src="./docs/docs/assets/logo.png" alt="Logo" width="128" height="128">
  </a>

  <!-- SKIPPER TLDR -->
  <h1 align="center">Skipper</h3>
  <p>Inventory based templated configuration library inspired by the kapitan project</p>
  </br>
</div>

> This is the **develop** branch of skipper, which differs heavily from the **main** branch.
>
> Skipper was initially developed as a POC. On this branch the new skipper will hopefully emerge.

# What is skipper?

Skipper is a library which helps you to manage complex configuration and enables
you to use your large data-set inside templates.
Having one - central - set of agnostic configuration files will make managing
your clusters, infrastrucutre stages, etc. much easier. You can rely on
the inventory of data, modify it to target-specific needs, use the data in
templates, and be sure that whatever you're generating is always in sync with your inventoyu.
Whether you generate only a single file, or manage multi-stage multi-region infrastructure deployments doesn't matter.
Skipper is a library which enables you to easily build your own - company or project specific - configuration management.

Skipper is heavily inspired by the [Kapitan](https://kapitan.dev/) project. The difference
is that skipper is a library which does not make any assumptions of your needs (aka. not opinionated).
This allows you for example to have way more control over how you want to process your inventory.

Skipper is not meant to be a _one-size-fits-all_ solution. The goal of Skipper is to enable
you to create the own - custom built - template and inventory engine, without having to do the heavy lifing.

# TODO

- [ ] Core Features
  - [ ] Data Abstraction Layer
  - [ ] New Class structure
  - [ ] Namespacing / Inventory
  - [ ] Project Configuration
  - [ ] Target + Target-Inventory
  - [ ] References (local, global, target)
  - [ ] Template Engine
- [ ] Plugin System
  - [ ] Hook System
  - Plugins
    - `env` to provide environment variable access in classes
    - `copy` to directly copy files into the target compile output
    - `undefined` to ensure class paths are defined once the target is loaded
    - `schema` to support class schema validation
    - `secret` for secret management
    - `validate` to provide custom validation logic
    - `go_schema` to generate go types based off the yaml files (and write them into go files)
    - `terraform_data` to generate a terraform data module with the inventory content
    - `include` to enable class-based includes with complex inheritance
- [ ] Basic
