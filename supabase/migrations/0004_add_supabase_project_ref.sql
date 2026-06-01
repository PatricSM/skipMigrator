-- Adds a column to surface the user's Supabase project ref so the UI can
-- render a tailored post-migration checklist (with deep-links to the panel).
alter table public.migrations
  add column if not exists supabase_project_ref text;
