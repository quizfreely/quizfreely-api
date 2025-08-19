-- migrate:up
alter table term_progress
    add column term_correct_count int not null default 0,
    add column term_incorrect_count int not null default 0,
    add column def_correct_count int not null default 0,
    add column def_incorrect_count int not null default 0;

-- migrate:down
