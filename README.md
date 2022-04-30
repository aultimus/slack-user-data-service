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
* Use of slack library rather than using REST API


## Notes
Slack codechallenge app has these permissions:
* View information about a user’s identity, granted by 1 team member
* View the name, email domain, and icon for workspaces a user is connected to, granted by 1 team member
* View people in a workspace, granted by 1 team member
* View email addresses of people in a workspace, granted by 1 team member
* View profile details about people in a workspace, granted by 1 team member
* Set a user’s presence, granted by 1 team member
* Edit a user’s profile information and status, granted by 1 team member

## TODO
* use connection pool?
