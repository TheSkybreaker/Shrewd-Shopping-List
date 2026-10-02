# Changelog

## Milestone 1: backend

- PocketBase v0.40.4 used as a Go framework, with one migration for the `users` extra fields, `lists`, `items` and `push_subscriptions`, their API rules and indexes.
- `vapid` command, environment loaded from an optional `.env`, static serving that keeps the query on the `/lista` redirect, and API rule tests.
