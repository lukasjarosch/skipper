# Inventory - Overview
The inventory is basically just a set of user-defined `yaml` files with some special rules attached.
These yaml files can be composed in order to produce a final inventory which you can use inside [templates](#templates) to compile your desired target infrastructure.

Below is an overview of all the components which make up Skipper, and how they might interact with one-another.

![overview](../../assets/images/overview.svg)

## Point of view
Whatever you do as part of your daily job, you most certainly are concerned with some sort of infrastructure - otherwhise you would not have found Skipper.
Depending on what level ov experience you have, you will most likely have seen the quickly changing tools in our industry.

Tools rise and go, but your infrastructure does not. Once you migrated all of your stuff to tool *X*, it becomes obsolete and you need to start migration again.
Idioms change and for some reason you always feel left behind, although you keep up-to-date to current CNCNF releases or other upcoming stuff.

#### What you really want to describe
No matter what tooling you're using. All you want to do is to describe you infrastructure as good as it gets.
And every few months/years you will need to adjust because the "standard tooling" has changed.
I mean, the CIDR ranges of your network(s) haven't changed, so why should you completely overhaul your IaC just to meet yet another "standard".

*That really sucks, doens't it?*

Maybe you want to be agnostic about the tools you use and offer to your customers. 
How about you just write down stuff which *actually* matters, without being dependant on any certain technology.
The stuff which actually matters is the essence of Skipper. You write down things - in the way you prefer - and produce something awesome with it.

*Well, I'm trying to sell you on yet another tool, how can you not be dependant on Skipper?*

What about the following thought: "All of your `x` environments are prepared and documented automatically, based on one single source of truth". 
That's what Skipper enables you to do, without throwing rocks in your way.

Skipper is just a way for you to reach your freedom. I don't care if you use Skipper - now or in the future. 
All I want to do is to allow you to escape the braces of modern cloud tooling.


#### Interesting thought, tell me more!

Here are the facts what skipper is about:

- Skipper wants you to abstract away all the information at your disposal - before thinking about the technologies used
- Infrastructure/Code/Documentation/Bootstraping/... is just an aggregation of information you should already have
- Skipper helps you to aggregate, cumulate and leverage your information to build the next big thing.
- Secret management should not be hard, it should be automatic. Skipper has got you covered!


> Skipper is a library which does the heavy lifting. Creating something awesome is up to you :wink:!
