To set up users BEFORE being able to run db migrations to setup the actual schema:
```sql
CREATE ROLE quizfreely_db_admin NOINHERIT LOGIN;

CREATE DATABASE quizfreely_db OWNER quizfreely_db_admin;

\connect quizfreely_db

CREATE ROLE quizfreely_api NOINHERIT LOGIN;
GRANT CONNECT ON DATABASE quizfreely_db TO quizfreely_api;

-- remember to set the admin/migration user's password
\password quizfreely_db_admin
-- remember to set the api user's password
\password quizfreely_api
```

if your/our database already exists with our `quizfreely_api` user, but no `quizfreely_db_admin` user:
```sql
CREATE ROLE quizfreely_db_admin NOINHERIT LOGIN;

ALTER DATABASE quizfreely_db OWNER TO quizfreely_db_admin;

ALTER SCHEMA auth OWNER TO quizfreely_db_admin;
ALTER TYPE auth_type_enum OWNER TO quizfreely_db_admin;
ALTER TABLE auth.users OWNER TO quizfreely_db_admin;
ALTER TABLE auth.sessions OWNER TO quizfreely_db_admin;
ALTER TABLE public.studysets OWNER TO quizfreely_db_admin;
ALTER INDEX textsearch_title_idx OWNER TO quizfreely_db_admin;
ALTER TABLE public.studyset_progress OWNER TO quizfreely_db_admin;
ALTER TABLE public.search_queries OWNER TO quizfreely_db_admin;
ALTER TABLE public.studyset_settings OWNER TO quizfreely_db_admin;

-- remember to set the admin/migration user's password
\password quizfreely_db_admin
```

## search-queries.sql

optionally, to populate `search_queries`, you can manually run `db/search-queries.sql`.
