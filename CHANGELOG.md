# Changelog

## Milestone 2: frontend base

- Static Astro frontend built into `backend/pb_public`, with Alpine stores, self-hosted fonts, the dark theme only, login with page guards, home, new list, list detail with add, check, delete and undo, and Vitest tests for the parser, dates and grouping.
- Realtime sync through the stores, reloaded when the app returns or the connection comes back; `items.added_at` puts an item taken back from the cart at the end of the items to buy.

## Milestone 1: backend

- PocketBase v0.40.4 used as a Go framework, with one migration for the `users` extra fields, `lists`, `items` and `push_subscriptions`, their API rules and indexes.
- `vapid` command, environment loaded from an optional `.env`, static serving that keeps the query on the `/lista` redirect, and API rule tests.
