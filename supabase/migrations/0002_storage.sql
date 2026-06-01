-- Storage buckets for source/output ZIPs.
-- Note: bucket creation can also be done via the dashboard or supabase CLI.
-- This migration is for reproducibility.

insert into storage.buckets (id, name, public)
values
    ('source-zips', 'source-zips', false),
    ('output-zips', 'output-zips', false)
on conflict (id) do nothing;

-- Policies: users can upload and read their own paths (prefix = user id).
-- Service role (used by the backend worker) bypasses RLS, so it can access all paths.

create policy "users upload to own source-zips prefix"
    on storage.objects for insert
    with check (
        bucket_id = 'source-zips'
        and split_part(name, '/', 1) = auth.uid()::text
    );

create policy "users read own source-zips"
    on storage.objects for select
    using (
        bucket_id = 'source-zips'
        and split_part(name, '/', 1) = auth.uid()::text
    );

create policy "users read own output-zips"
    on storage.objects for select
    using (
        bucket_id = 'output-zips'
        and split_part(name, '/', 1) = auth.uid()::text
    );
