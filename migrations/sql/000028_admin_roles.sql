-- +goose Up

-- 1. Insert the 5 granular admin roles into your existing auth.roles table
INSERT INTO auth.roles (name, description) VALUES 
    ('SUPER_ADMIN', 'Full access to all platform features, settings, and admin management'),
    ('OPERATIONS_ADMIN', 'Manages live operations, trip interventions, and driver zones'),
    ('FINANCE_ADMIN', 'Manages driver payouts, financial reports, and promotions'),
    ('SUPPORT_ADMIN', 'Handles customer support tickets, refunds, and user disputes'),
    ('SAFETY_ADMIN', 'Monitors live SOS alerts and handles safety interventions')
ON CONFLICT (name) DO NOTHING;

-- 2. Promote any existing generic 'ADMIN' users to 'SUPER_ADMIN' 
-- so they don't lose their access when we lock down the routes later.
-- (We only run this if a generic 'ADMIN' role actually exists)
UPDATE auth.user_roles 
SET role_id = (SELECT id FROM auth.roles WHERE name = 'SUPER_ADMIN')
WHERE role_id = (SELECT id FROM auth.roles WHERE name = 'ADMIN');

-- +goose Down

DELETE FROM auth.roles WHERE name IN (
    'SUPER_ADMIN',
    'OPERATIONS_ADMIN',
    'FINANCE_ADMIN',
    'SUPPORT_ADMIN',
    'SAFETY_ADMIN'
);