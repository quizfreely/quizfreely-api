-- migrate:up
CREATE TYPE answer_with_enum AS ENUM (
    'TERM',
    'DEF'
);

CREATE TABLE term_confusion_pairs (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references auth.users (id) on delete cascade,
    term_id uuid not null references terms (id) on delete cascade,
    confused_term_id uuid not null references terms (id) on delete cascade,
    answered_with answer_with_enum not null,
    confused_count int,
    last_confused_at timestamptz not null default now()
);

grant select on term_confusion_pairs to quizfreely_api;
grant insert on term_confusion_pairs to quizfreely_api;
grant update on term_confusion_pairs to quizfreely_api;
grant delete on term_confusion_pairs to quizfreely_api;

-- migrate:down
