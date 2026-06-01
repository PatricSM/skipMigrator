-- skip-migrator initial schema.
-- Users come from auth.users (Supabase Auth).
-- Migrations records each Lovable→Skip job a user kicked off.

create extension if not exists "uuid-ossp";

create table if not exists public.migrations (
    id                  uuid primary key default uuid_generate_v4(),
    user_id             uuid not null references auth.users(id) on delete cascade,
    status              text not null default 'queued'
                          check (status in ('queued','running','success','failed')),
    source_zip_path     text not null,           -- key in `source-zips` bucket
    output_zip_path     text,                    -- key in `output-zips` bucket (set on success)
    build_log           text,                    -- captured pnpm install/build output
    error_message       text,                    -- short description on failure
    pixel_perfect       boolean not null default false,
    validate            boolean not null default true,
    supabase_strategy   text not null default 'extract'
                          check (supabase_strategy in ('extract','new')),
    created_at          timestamptz not null default now(),
    updated_at          timestamptz not null default now()
);

create index if not exists migrations_user_id_created_at_idx
    on public.migrations (user_id, created_at desc);
create index if not exists migrations_status_idx
    on public.migrations (status)
    where status in ('queued','running');

-- RLS: users can read/insert their own; service role bypasses.
alter table public.migrations enable row level security;

create policy migrations_user_select on public.migrations
    for select using (auth.uid() = user_id);
create policy migrations_user_insert on public.migrations
    for insert with check (auth.uid() = user_id);

-- Trigger to keep updated_at fresh
create or replace function public.set_updated_at()
returns trigger language plpgsql as $$
begin
    new.updated_at := now();
    return new;
end$$;

drop trigger if exists migrations_set_updated_at on public.migrations;
create trigger migrations_set_updated_at
    before update on public.migrations
    for each row execute function public.set_updated_at();

-- Enable realtime on the migrations table so the frontend can subscribe.
-- (Supabase requires the table be in the supabase_realtime publication.)
alter publication supabase_realtime add table public.migrations;
