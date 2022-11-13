---
hide:
  - navigation
  - toc
---
![](./assets/logo.png){ width="100"}
# Skipper 

<a href="https://pkg.go.dev/github.com/lukasjarosch/skipper"><img src="https://pkg.go.dev/badge/github.com/lukasjarosch/skipper.svg" alt="Go Reference"></a>
<a href="https://goreportcard.com/report/github.com/lukasjarosch/skipper"><img src="https://goreportcard.com/badge/github.com/lukasjarosch/skipper"></a>
<a href="https://github.com/lukasjarosch/skipper/actions/workflows/test.yml"><img src="https://github.com/lukasjarosch/skipper/actions/workflows/test.yml/badge.svg"></a>


## What is Skipper?

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


## Values

- Skipper wants you to abstract away all the information at your disposal - before thinking about the technologies used
- Infrastructure/Code/Documentation/Bootstraping/... is just an aggregation of information you should already have
- Skipper helps you to aggregate, cumulate and leverage your information to build the next big thing.
- Secret management should not be hard, it should be automatic. Skipper has got you covered!


> Skipper is a library which does the heavy lifting. Creating something awesome is up to you :wink:!
