<div align="center">
  <a href="https://github.com/lukasjarosch/skipper">
    <img src="./assets/logo.png" alt="Logo" width="128" height="128">
  </a>

<h1 align="center">Skipper</h3>
<p>Inventory based templated configuration library based on the kapitan project</p>
</br>
</div>

# What is skipper?

Skipper is heavily inspired by the awesome [Kapitan](https://kapitan.dev/) project. The major difference
is that skipper is primarily meant to be used as library. This allows you for example to have way more
control over the data-structures used inside the templates. 

## Roadmap

- [x] Allow self referencing within classes
  - When writing classes one will very likely need to reference a valuie of the same class somewhere else. 
    Think about defining a docker image once and reusing it throughout the class, but with different tags.
  - Enable the notation `${object:key}` within classes and targets
- [x] Introduce a default set of `${variable}` variables to be used within targets and classes
  - First candidade is to have `${target_name}` accessible everywhere
- [ ] Allow definition of custom variables within classes
- [ ] Class inheritance. Currently only targets can `use` classes but it would be nice if classes could also use different classes
  - This would introduce a higher level of inheritance which users can set-up for their inventory.
- [ ] `<no value>` detection in rendered templates
  - If for some reason a template uses a value which is not set, the user should have the ability to detect that post generation.
  - Introduce a verify mechanism which ideally checks for missing values and maybe also extracts which template-key it originated from
  - This can be evaluated before the template files are written to disk
  - We can have a strict mode which fails on these errors, or not
- [x] Add the option to inject arbitrary maps into the inventory with custom keys (`inventory.AddKey(key string, data Data)`)
  - This is very useful if you have a different data-structure which you want to add
  - For example if your app has a model which can be written via an HTTP API, you might want to be able to use these data as well


# Documentation
> This documentation is work in progress and will be moved to it's own place in the future. 

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
