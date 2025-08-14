To set up users BEFORE being able to run db migrations to setup the actual schema:
```sql
CREATE ROLE quizfreely_db_admin LOGIN PASSWORD 'REAL_PASSWORD_GOES_HERE';

CREATE DATABASE quizfreely_db OWNER quizfreely_db_admin;

\connect quizfreely_db

CREATE ROLE quizfreely_api NOINHERIT LOGIN;
GRANT CONNECT ON DATABASE quizfreely_db TO quizfreely_api;
```

if your/our database already exists with our `quizfreely_api` user, but no `quizfreely_db_admin` user:
```sql
CREATE ROLE quizfreely_db_admin LOGIN PASSWORD 'REAL_PASSWORD_GOES_HERE';

ALTER DATABASE quizfreely_db OWNER TO quizfreely_db_admin;
```
