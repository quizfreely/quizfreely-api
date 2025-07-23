# Quizfreely's API Rewritten in Go

**Abstract**

A free and open source studying tool runs on **1GB RAM**.

Two NodeJS server applications share the same server: a GraphQL API using Fastify with a PostgreSQL database and a web app using SvelteKit.

So what if we speedrun rewriting that GraphQL API in Golang instead of JavaScript? A JS engine like NodeJS will inherently use more resources than a compiled language like Go. If we want to save a few MBs of RAM, why not spend ~~another year~~ *a few hours\** rewriting the entire API?

