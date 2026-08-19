-- +goose Up
-- Driver Profiles (1-to-1 with auth.users)
CREATE TABLE auth.driver_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    license_number VARCHAR(100),
    license_expiry DATE,
    onboarding_step VARCHAR(50) NOT NULL DEFAULT 'PROFILE_SETUP',
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_driver_status CHECK (status IN ('PENDING', 'ACTIVE', 'SUSPENDED', 'BANNED')),
    CONSTRAINT valid_onboarding_step CHECK (onboarding_step IN ('PROFILE_SETUP', 'VEHICLE_DETAILS', 'DOCUMENTS', 'IDENTITY_CHECK', 'ADMIN_REVIEW', 'COMPLETED'))
);

-- Vehicles (1-to-many with driver_profiles)
CREATE TABLE auth.vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_profile_id UUID NOT NULL REFERENCES auth.driver_profiles(id) ON DELETE CASCADE,
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INT NOT NULL,
    color VARCHAR(50) NOT NULL,
    plate_number VARCHAR(50) NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL DEFAULT 'SEDAN',
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_vehicle_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    CONSTRAINT valid_vehicle_type CHECK (vehicle_type IN ('SEDAN', 'SUV', 'VAN', 'LUXURY', 'MOTORCYCLE'))
);

-- Driver Documents (Personal & Vehicle)
CREATE TABLE auth.driver_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_profile_id UUID NOT NULL REFERENCES auth.driver_profiles(id) ON DELETE CASCADE,
    vehicle_id UUID REFERENCES auth.vehicles(id) ON DELETE CASCADE, -- Nullable: Null if it's a personal ID/License
    doc_type VARCHAR(50) NOT NULL,
    file_key VARCHAR(255) NOT NULL, -- S3 object key
    file_url VARCHAR(512) NOT NULL, -- Presigned or public URL
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    admin_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_doc_type CHECK (doc_type IN (
        'SA_ID_OR_PASSPORT',
        'DRIVER_LICENSE',
        'PRDP',
        'PROFILE_PHOTO',
        'BACKGROUND_CHECK',
        'VEHICLE_REGISTRATION',
        'VEHICLE_ROADWORTHY',
        'VEHICLE_INSURANCE',
        'OPERATING_LICENSE',
        'VEHICLE_INSPECTION',
        'VEHICLE_PHOTOS'
    )),
    CONSTRAINT valid_doc_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
);

-- Identity Verifications (KYC Provider Tracking)
CREATE TABLE auth.identity_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_profile_id UUID NOT NULL REFERENCES auth.driver_profiles(id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL,
    provider_verification_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    webhook_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_verification_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'ERROR'))
);

-- Document Requirements (Admin Configuration for Mandatory vs Optional)
CREATE TABLE auth.document_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doc_type VARCHAR(50) UNIQUE NOT NULL,
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_vehicle BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed the default requirements based on South African e-hailing regulations
INSERT INTO auth.document_requirements (doc_type, is_mandatory, applies_to_vehicle, description) VALUES
('SA_ID_OR_PASSPORT', TRUE, FALSE, 'South African ID or valid passport'),
('DRIVER_LICENSE', TRUE, FALSE, 'Valid driver''s licence'),
('PRDP', TRUE, FALSE, 'Professional Driving Permit'),
('PROFILE_PHOTO', TRUE, FALSE, 'Driver profile photo'),
('BACKGROUND_CHECK', TRUE, FALSE, 'Criminal/background check'),
('VEHICLE_REGISTRATION', TRUE, TRUE, 'Vehicle registration/licence documents'),
('VEHICLE_ROADWORTHY', TRUE, TRUE, 'Vehicle roadworthy certificate'),
('VEHICLE_INSURANCE', TRUE, TRUE, 'Vehicle insurance'),
('OPERATING_LICENSE', FALSE, TRUE, 'Operating licence / e-hailing permit documentation, where applicable'),
('VEHICLE_INSPECTION', FALSE, TRUE, 'Vehicle inspection, where required'),
('VEHICLE_PHOTOS', TRUE, TRUE, 'Vehicle pictures')
ON CONFLICT (doc_type) DO NOTHING;

-- Indexes
CREATE INDEX idx_driver_profiles_user_id ON auth.driver_profiles(user_id);
CREATE INDEX idx_driver_profiles_status ON auth.driver_profiles(status);
CREATE INDEX idx_vehicles_driver_id ON auth.vehicles(driver_profile_id);
CREATE INDEX idx_driver_documents_profile_id ON auth.driver_documents(driver_profile_id);
CREATE INDEX idx_identity_verifications_profile_id ON auth.identity_verifications(driver_profile_id);
CREATE INDEX idx_document_requirements_type ON auth.document_requirements(doc_type);

-- +goose Down
DROP TABLE IF EXISTS auth.document_requirements;
DROP TABLE IF EXISTS auth.identity_verifications;
DROP TABLE IF EXISTS auth.driver_documents;
DROP TABLE IF EXISTS auth.vehicles;
DROP TABLE IF EXISTS auth.driver_profiles;