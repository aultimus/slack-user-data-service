Please use this README file for notes of any design decisions, trade-offs, or improvements you’d make to the project.

Please see the [instructions](INSTRUCTIONS.md) to get started.

## Running
Put environment variables in a `dev.env` file at the top level of the project,
docker-compose will look for this file.

The `SLACK_API_TOKEN` and `SLACK_VERIFICATION_TOKEN` environment variables are
a prerequisite for running this service. `SLACK_API_TOKEN` will need to be a
slack token with `users:read` scope. For this specific app these values can
be found [here](https://api.slack.com/apps/A03CYL14A5B)

In order to run in development mode execute:
`make run`

This will spin up the app and an accompanying database via docker-compose.

The service can be accessed on port `3000` with a web browser, so running locally
the users endpoint can be accessed at `http://localhost:3000/users`

## Testing
In order to run the integration tests execute:
`make integrationtest`
This command will return a positive exit code if the tests fail so is easily
usable in CI. The integration test runs a server to mock the slack api, sends
the app events and makes requests to the app endpoint to verify behaviour.

## Tradeoffs
* To use an ORM or not to use an ORM? I considered using an ORM but wished to
keep the code more flexible to allow for extensions to the service. We could always
introduce one if maintaining the DB code becomes too much of a burden, as it
stands sqlx saves a bunch of legwork.
* We use the slack library `github.com/slack-go/slack` rather than using the REST
API. The slack library provides us with some predefined types and adds some nice
features such as out of the box pagination and verification. It does add some
complexity into the code in that we need to deal with more types and cannot simply
treat the slack response as raw json but hopefully it provides safety in its stead
and reliability in the face of any api changes.
* I wrote integration tests as I wanted to test the full surface area of the
service. This is a trade off against writing unit tests which would be quicker to
write and run but would test less surface area.
* I chose not to process team_joined events but it would not be much work to add
this functionality, this is because I have observed user_changed events to
accompany team_join events in every occasion that I have witnessed them.
* The integration tests currently parse the html form served on /users, this code
is brittle to changes in format / extensions but provides good confidence that the
user facing table feature which was the deliverable for this assignment is
functioning as expected. It would be more stable to test against a REST interface
that returned all users as JSON, though this would not test the user facing table,
the tests could fairly easily be adapted to test such an interface.
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
* To pull manually from slack api use: `curl -X POST -H "Authorization: Bearer $SLACK_API_TOKEN" https://slack.com/api/users.list | python3 -m json.tool`
## TODO
* Add unit tests that provide quick feedback on regressions, also test failure cases such as DB being down
* Configure proper db password
