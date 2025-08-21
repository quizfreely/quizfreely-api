-- migrate:up
CREATE TABLE practice_tests (
    id uuid primary key default gen_random_uuid(),
    timestamp timestamptz not null default now(),
    user_id uuid not null references auth.users (id) on delete cascade,
    studyset_id uuid not null references studysets (id) on delete cascade,
    questions_correct smallint,
    questions_total smallint,
    questions jsonb
);

grant select on practice_tests to quizfreely_api;
grant insert on practice_tests to quizfreely_api;
grant update on practice_tests to quizfreely_api;
grant delete on practice_tests to quizfreely_api;

-- migrate:down
