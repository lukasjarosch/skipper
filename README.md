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

Skipper is not meant to be a *one-size-fits-all* solution. The goal of Skipper is to enable
you to create the own - custom built - template and inventory engine, without having to do the heavy lifing.

### Acknowledgments
- Logo: <a href="https://www.flaticon.com/de/kostenlose-icons/kapitan" title="kapitän Icons">Skipper Logo designed by freepik - Flaticon</a>

