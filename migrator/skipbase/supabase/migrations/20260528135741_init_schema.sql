-- Create profiles table
CREATE TABLE IF NOT EXISTS public.profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    full_name TEXT,
    avatar_url TEXT,
    role TEXT DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

-- Policies
DROP POLICY IF EXISTS "Users can view own profile" ON public.profiles;
CREATE POLICY "Users can view own profile" ON public.profiles
    FOR SELECT TO authenticated USING (auth.uid() = id);

DROP POLICY IF EXISTS "Users can update own profile" ON public.profiles;
CREATE POLICY "Users can update own profile" ON public.profiles
    FOR UPDATE TO authenticated USING (auth.uid() = id);

-- Trigger for auto-profile creation
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
    INSERT INTO public.profiles (id, full_name)
    VALUES (NEW.id, NEW.raw_user_meta_data->>'full_name');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Seed User
DO $$
DECLARE
    new_user_id uuid;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM auth.users WHERE email = 'patric.martins@adapta.org') THEN
        new_user_id := gen_random_uuid();
        INSERT INTO auth.users (
            id, instance_id, email, encrypted_password, email_confirmed_at,
            created_at, updated_at, raw_app_meta_data, raw_user_meta_data,
            is_super_admin, role, aud,
            confirmation_token, recovery_token, email_change_token_new,
            email_change, email_change_token_current,
            phone, phone_change, phone_change_token, reauthentication_token
        ) VALUES (
            new_user_id,
            '00000000-0000-0000-0000-000000000000',
            'patric.martins@adapta.org',
            crypt('Skip@Pass', gen_salt('bf')),
            NOW(), NOW(), NOW(),
            '{"provider": "email", "providers": ["email"]}',
            '{"full_name": "Patric Martins"}',
            false, 'authenticated', 'authenticated',
            '', '', '', '', '',
            NULL, '', '', ''
        );
        
        INSERT INTO public.profiles (id, full_name, role)
        VALUES (new_user_id, 'Patric Martins', 'admin')
        ON CONFLICT (id) DO NOTHING;
    END IF;
END $$;
