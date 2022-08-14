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


### Acknowledgments
- Logo: <a href="https://www.flaticon.com/de/kostenlose-icons/kapitan" title="kapitän Icons">Skipper Logo designed by freepik - Flaticon</a>
