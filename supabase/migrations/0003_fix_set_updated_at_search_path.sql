-- Resolve linter warning 0011_function_search_path_mutable
-- by fixing the search_path on public.set_updated_at.
create or replace function public.set_updated_at()
returns trigger
language plpgsql
security definer
set search_path = public, pg_temp
as $$
begin
    new.updated_at := now();
    return new;
end$$;
