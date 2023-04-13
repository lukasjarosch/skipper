Dynamic variables are references into your data inventory.
They are immensly useful to compose data into more complex values without having to redefine stuff all over the place.

## Local-referencing variables

```yaml title="myClass.yaml"
myClass:
    foo: bar
    bar: 
        - baz
        - world
    hello: ${bar:1} 
```


## Absolute-referencing variables
