Please use this README file for notes of any design decisions, trade-offs, or improvements you’d make to the project.

Please see the [instructions](INSTRUCTIONS.md) to get started.

## Running
Put environment variables in a `dev.env` file at the top level of the project,
docker-compose will look for this file.

The `SLACK_API_TOKEN` environment variable is a prerequisite for running
this service. This will need to be a slack token with `users:read` scope.

In order to run in development mode execute:
`make run`

This will spin up the app and an accompanying database via docker-compose.

## Testing
In order to run the tests execute:
TODO

## Tradeoffs
* To use an ORM or not to use an ORM? I considered using an ORM but have not
used one recently in golang and wished to keep the code more flexible to allow
for extensions to the service. We could always introduce one if maintaining the
DB code becomes too much of a burden, as it stands sqlx saves a bunch of legwork.
* We use the slack library `github.com/slack-go/slack` rather than using the REST
API. The slack library provides us with some predefined types and adds some nice
features such as out of the box pagination and verification. It does add some
complexity into the code in that we need to deal with more types and cannot simply
treat the slack response as raw json but hopefully it provides safety in its stead
and reliability in the face of any api changes.

## Notes
Slack codechallenge app has these permissions:
    * View information about a user’s identity, granted by 1 team member
    * View the name, email domain, and icon for workspaces a user is connected to, granted by 1 team member
    * View people in a workspace, granted by 1 team member
    * View email addresses of people in a workspace, granted by 1 team member
    * View profile details about people in a workspace, granted by 1 team member
    * Set a user’s presence, granted by 1 team member
    * Edit a user’s profile information and status, granted by 1 team member
* SQLX provides a connection pool for us

## TODO
* Test deactivating / deleting user
