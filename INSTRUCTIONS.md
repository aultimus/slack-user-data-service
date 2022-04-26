# Setup

- An ngrok tunnel has been setup for you at https://matthew-ault.interview.workos.dev. Use the provided `ngrok.yml` config file to start it.
- We’ve created a Slack workspace for development: https://matthew-ault-workos.slack.com. You’ll receive an invite for this workspace via email. Feel free to use this space as a sandbox—add users, modify users, etc.
- We’ve also created a pre-configured Slack App that will send events to /webhooks on your ngrok tunnel. After you create your Slack account, you can view the app configuration here: https://api.slack.com/apps/A03CYL14A5B/event-subscriptions.

Let us know if you have any questions or want to jump on a quick call. We also enabled Slack Connect in the #workos-team channel if you need to ping our Engineering team.

# The Challenge

Build a web application that syncs users from a Slack workspace and displays the user list.

## How it should work

### Requirements

- User updates (profile updates, account deactivation, etc) made through Slack should be reflected in your user records in the database.
- Newly added users should also be reflected in the database.
- A table of users should be viewable via a UI.
- Existing users from the workspace should be persisted to a database.

We recommend persisting the following user fields: id, name, deleted, real_name, tz, profile object (status_text, status_emoji, image_512).

## How to build it

### Initial Setup

- We’ll provide a pre-configured Slack workspace that’s populated with test data.
- For development, you’ll need to use ngrok to receive Slack webhooks.
- Pick a language and a setup that you know well. While there are obviously some constraints around what will work, we don't recommend picking technologies with which you're completely unfamiliar. This will make it easier for you to show off your skills as we go through the interview process!
- Use a **SQL-based database** to store data.

## Deliverables

1. URL for your app.
2. Use the `README.md` file with notes of any design decisions, trade-offs, or improvements you’d make.
3. Open a pull-request against the provided github repository with your changes.

## What we look for

This challenge is meant to help us see your best code, and to showcase your judgment. When we evaluate the challenge, we look at how focused you were on meeting the requirements, at the simplicity and correctness of your architecture, at your use of appropriate design patterns, your choice of data structures, and your use of best practices.
