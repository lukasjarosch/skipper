A class is just any arbitrary yaml file with a few rules attached.
For the most part you are free to do whatever you want inside a class.

Think of classes as building-blocks which describe different aspects of your project.

#### Class Rules
 In order for your inventory to be maintainable by Skipper, there are a few - very basic - rules which you need to comply.

1. There can only be **one root key** inside the class
2. The root key must match the filename of your class. A class `project.yaml` is *expected* to use `project` as root key.
3. The class **cannot be called** `target.yaml`

#### Example Use-case
You got yourself a little project where you need to setup some cloud resources, nice!

- You want to use terraform (at least for now, maybe you need to switch in the future)
- You want to store the terraform state remotely
- You want to have a script which bootstraps the storage of your state
- You need various deployments of the same infrastructure (dev, staging, production, qa, ...)

We're not going down the rabbit-hole why generating terraform code is inhertently a better idea than to write it manually, just roll with it for now.
If you're really curious, check out [this short summary by Eden Reich](https://www.eden-reich.com/engineering-blog/infrastructure-as-data/).
