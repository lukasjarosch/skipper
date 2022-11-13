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

# Core Concepts
Skipper has a few concepts, but not all of them are necessary to understand how Skipper works.
More in-depth informatation about Skippers concepts can be found [in our docs](https://lukasjarosch.github.io/skipper/concepts/).

## Inventory
The inventory is the heart of every Skipper-enabled project. It is your data storage, the single source of truth.
It is a user-defined collection of YAML files (classes and targets).

### Classes
Classes are YAML files in which you can define information about every part of your project.
These classes become your building blocks and therefore the heart of your project.

### Targets
A target represents an instance of your project. Targets are defined with YAML files as well.
They use skipper-keywords to *includ* classes, relevant for that instance.
Inside a target config you are also able to overwrite any kind of information (change the location in which your resources are deployed for example).

## Templates
Templates (Skipper is using [go templates](https://pkg.go.dev/text/template)) have access to your target and classes.
You can build generic templates and aggregate your data into it, without having to re-write files for different stages.
Having a documentation, specific to an instance (stage) of your project, can be quite useful and is easy to implement with Skipper.


# Idea collection

- [ ] Allow static file copying instead of rendering it as template (e.g. copy a zip file from templates to compiled)
- [ ] Add timing stats (benchmark, 'compiled in xxx') to compare with kapitan
- [ ] Class inheritance. Currently only targets can `use` classes but it would be nice if classes could also use different classes
  - This would introduce a higher level of inheritance which users can set-up for their inventory.


# Documentation
> This documentation is work in progress and will be moved to it's own place in the future. 

## Classes
A class is a yaml file which defines arbitrary information about your project.

There is only two rules for classes:
  - The filename of the class must be the root key of the yaml struct
  - The filename **cannot** be `target.yaml`, resulting in a root key of `target`.
    - Although this will not return an error, you simply will not be able to use the class as the actual target will overwrite it completely.

This means that if your class is called `pizza.yaml`, the class must look like this:

```yaml
pizza:
  # any value
```

## Targets
A target usually is a speparate environment in your infrastructure or a single namespace in your Kubernetes cluster.
Targets `use` classes to pull in the required innventory data in order to produce the correct tree which is required in order to render the templates.

On any given run, Skipper only allows to set **one** target. This is to ensure that the generated map of data is consistent.

The way a target makes uses of the inventory is by using the `use` clause which tells Skipper which classes to include in the assembly of the target inventory.


### Naming

The name of the target is given by its filename. So if your target is called `development.yaml`, the target name will be `development`. 

The structure of a target file is pretty simple, there are only two rules:

- The top-level key of the target **must** be `target`
- There must be a key `target.use` which *has to be* an array and tells Skipper which classes this particular target requires.

Below you'll find the most basic example of a target.
The target does not define values itself, it just uses values from a class `project.common`.
The class must be located in the `classPath` passed into `Inventory.Load()`, where path separators are replaced by a dot.

So if your classPath is `./inventory/classes`, referencing `foo.bar` will make Skipper attempt to load `./inventory/classes/foo/bar.yaml`.

```yaml
target:
  use:
    project.common
```

## Variables
Variables in Skipper always have the same format: `${variable_name}` 

Skipper has *three* different types of variables.

1. [Dynamic Variables](#dynamic-variables)
2. [Predefined Variables](#predefined-variables)
3. [User-defined Variables](#user-defined-variables)

### Dynamic Variables

Dynamic variables are variables which use a *selector path* to point to existing values which are defined in your inventory.

Consider the following class **images.yaml**
```yaml
images:
  base_image: some/image

  production:
    image: ${images:base_image}:v1.0.0
  staging:
    image: ${images:base_image}:v2.0.0-rc1
  development:
    image: ${images:base_image}:latest
```

Once the class is processed, the class looks like this:
```yaml
images:
  base_image: some/image

  production:
    image: some/image:v1.0.0
  staging:
    image: some/image:v2.0.0-rc1
  development:
    image: some/image:latest
```

The name of the variable uses common *dot-notation*, except that we're using ':' instead of dots.
We chose to use colons because they are easier to read inside the curly braces.

### Predefined Variables

Predefined variables could also be considered constants - at least from a user perspective.
The predefined variables can easily be defined as `map[string]interface{}`, where the keys are
the variable names.

You have to pass your predefined variables to the `Inventory.Data()` call, then they are evaluated.
If you do not pass these variables, the function will return an error as it will attempt to treat them as dynamic variables.

Consider the following example (your main.go)
```go
// ...

predefinedVariables := map[string]interface{}{
    "target_name": "develop",
    "output_path": "foo/bar/baz",
    "company_name": "AcmeCorp"
}

data, err := inventory.Data(target, predefinedVariables)
if err != nil {
    panic(err)
}

// ...
```

You will now be able to use the variables `${target_name}`, `${output_path}` and `${company_name}`
throughout your yaml files and have Skipper replace them dynamically once you call `Data()` on the inventory.

### User-defined Variables

**TODO**

### Acknowledgments
- Logo: <a href="https://www.flaticon.com/de/kostenlose-icons/kapitan" title="kapitän Icons">Skipper Logo designed by freepik - Flaticon</a>

