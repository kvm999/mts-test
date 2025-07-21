package shared

import (
	"errors"
)

var (
	ErrMigrationDirectoryNotFound = errors.New("migration directory not found")

	ErrLicenseValidation   = errors.New("license validation")
	ErrLicenseVerification = errors.New("license verification")
)
