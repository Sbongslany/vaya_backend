package domain

type DriverStatus string

const (
	DriverStatusPending   DriverStatus = "PENDING"
	DriverStatusActive    DriverStatus = "ACTIVE"
	DriverStatusSuspended DriverStatus = "SUSPENDED"
	DriverStatusBanned    DriverStatus = "BANNED"
)

type OnboardingStep string

const (
	StepProfileSetup   OnboardingStep = "PROFILE_SETUP"
	StepVehicleDetails OnboardingStep = "VEHICLE_DETAILS"
	StepDocuments      OnboardingStep = "DOCUMENTS"
	StepIdentityCheck  OnboardingStep = "IDENTITY_CHECK"
	StepAdminReview    OnboardingStep = "ADMIN_REVIEW"
	StepCompleted      OnboardingStep = "COMPLETED"
)

type VehicleType string

const (
	VehicleTypeSedan      VehicleType = "SEDAN"
	VehicleTypeSUV        VehicleType = "SUV"
	VehicleTypeVan        VehicleType = "VAN"
	VehicleTypeLuxury     VehicleType = "LUXURY"
	VehicleTypeMotorcycle VehicleType = "MOTORCYCLE"
)

type DocumentType string

const (
	// --- Personal Documents ---
	DocTypeIDOrPassport    DocumentType = "SA_ID_OR_PASSPORT"
	DocTypeDriverLicense   DocumentType = "DRIVER_LICENSE"
	DocTypePrDP            DocumentType = "PRDP"
	DocTypeProfilePhoto    DocumentType = "PROFILE_PHOTO"
	DocTypeBackgroundCheck DocumentType = "BACKGROUND_CHECK"

	// --- Vehicle Documents ---
	DocTypeVehicleRegistration DocumentType = "VEHICLE_REGISTRATION"
	DocTypeRoadworthy          DocumentType = "VEHICLE_ROADWORTHY"
	DocTypeVehicleInsurance    DocumentType = "VEHICLE_INSURANCE"
	DocTypeOperatingLicense    DocumentType = "OPERATING_LICENSE"
	DocTypeVehicleInspection   DocumentType = "VEHICLE_INSPECTION"
	DocTypeVehiclePhotos       DocumentType = "VEHICLE_PHOTOS"
)

type DocumentStatus string

const (
	DocStatusPending  DocumentStatus = "PENDING"
	DocStatusApproved DocumentStatus = "APPROVED"
	DocStatusRejected DocumentStatus = "REJECTED"
)

type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "PENDING"
	VerificationApproved VerificationStatus = "APPROVED"
	VerificationRejected VerificationStatus = "REJECTED"
	VerificationError    VerificationStatus = "ERROR"
)