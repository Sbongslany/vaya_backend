-- +goose Up
INSERT INTO auth.roles (name, description) VALUES
    ('PASSENGER', 'Standard passenger account'),
    ('DRIVER', 'Verified driver account'),
    ('ADMIN', 'Standard administrator'),
    ('SUPER_ADMIN', 'Super administrator with full system access'),
    ('SUPPORT_ADMIN', 'Customer support administrator'),
    ('SAFETY_ADMIN', 'Safety and incident response administrator'),
    ('FINANCE_ADMIN', 'Finance and payouts administrator')
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DELETE FROM auth.roles WHERE name IN (
    'PASSENGER', 'DRIVER', 'ADMIN', 'SUPER_ADMIN', 
    'SUPPORT_ADMIN', 'SAFETY_ADMIN', 'FINANCE_ADMIN'
);