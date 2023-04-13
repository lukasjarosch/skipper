Variables allow you to use data which is not (yet) defined or reference data which is defined somewhere else.

In Skipper, there are three types of variables:

1. [Dynamic Variables](./dynamic.md) are used to reference data in the inventory.
1. [Static Variables](./static.md) are built into your binary which uses skipper, hence they are known at compile-time.
1. [User-Defined Variables](./user-defined.md) are defined and referenced custom values in the inventory.

## Format
The format of variables is `${variable_name}`. It can be used in [Classes](../classes.md) and [Targets](../targets.md).

> The regex used to match variables is: `\$\{((\w*)(\:\w+)*)\}`
