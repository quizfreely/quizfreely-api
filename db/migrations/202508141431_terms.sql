-- migrate:up
create table terms (
    id uuid primary key default gen_random_uuid(),
    term text,
    def text,
    studyset_id uuid not null references studysets (id) on delete cascade,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

grant select on terms to quizfreely_api;
grant insert on terms to quizfreely_api;
grant update on terms to quizfreely_api;
grant delete on terms to quizfreely_api;

create table term_progress (
    id uuid primary key default gen_random_uuid(),
    term_id uuid not null references terms (id) on delete cascade,
    user_id uuid not null references auth.users (id) on delete cascade,
    term_first_reviewed_at timestamptz,
    term_last_reviewed_at timestamptz,
    term_review_count int,
    def_first_reviewed_at timestamptz,
    def_last_reviewed_at timestamptz,
    def_review_count int,
    term_leitner_system_box smallint,
    def_leitner_system_box smallint
);

grant select on term_progress to quizfreely_api;
grant insert on term_progress to quizfreely_api;
grant update on term_progress to quizfreely_api;
grant delete on term_progress to quizfreely_api;

INSERT INTO terms (term, def, studyset_id)
SELECT
    elem->>0 AS term,
    elem->>1 AS def,
    s.id AS studyset_id
FROM studysets s
CROSS JOIN LATERAL jsonb_array_elements(s.data->'terms') AS elem;

ALTER TABLE studysets
DROP COLUMN data;

DROP TABLE studyset_progress;

DROP TABLE studyset_settings;

-- migrate:down
