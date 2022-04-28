Please use this README file for notes of any design decisions, trade-offs, or improvements you’d make to the project.

Please see the [instructions](INSTRUCTIONS.md) to get started.


## Tradeoffs
* To use an ORM or not to use an ORM? I considered using an ORM but have not
used one recently in golang and wished to keep the code more flexible to allow
for extensions to the service. We could always introduce one if maintaing DB
objects becomes too much of a burden
* We could create seperate db model structs, input structs and output api structs
it would certainly make sense to do this in a large project but this may hinder
speed of iteration when developing on this service at this stage of development
