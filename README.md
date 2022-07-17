<div align="center">
  <a href="https://github.com/lukasjarosch/skipper">
    <img src="./assets/logo.png" alt="Logo" width="128" height="128">
  </a>

<h3 align="center">Skipper</h3>
</div>

# What is skipper?

Skipper is heavily inspired by the awesome [Kapitan](https://kapitan.dev/) tool. The major difference
is that skipper is primarily meant to be used as library. This allows you for example to have way more
control over the data-structures used inside the templates. 

## Roadmap

- [ ] `<no value>` detection in rendered templates
  - If for some reason a template uses a value which is not set, the user should have the ability to detect that post generation.
  - Introduce a verify mechanism which ideally checks for missing values and maybe also extracts which template-key it originated from
- [ ] Add the option to inject arbitrary maps into the inventory with custom keys (`inventory.AddKey(key string, data Data)`)
  - This is very useful if you have a different data-structure which you want to add
  - For example if your app has a model which can be written via an HTTP API, you might want to be able to use these data as well


### Acknowledgments
- Logo: <a href="https://www.flaticon.com/de/kostenlose-icons/kapitan" title="kapitän Icons">Skipper Logo designed by freepik - Flaticon</a>
