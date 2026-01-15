# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [v1.2.0] - 2026-01-15

### Added
- New `FromPtrOrZero` function for safe pointer dereferencing with zero value fallback

## [v1.1.0] - 2025-12-25

### Added
- New `Dyn` type for dynamic/any values with full JSON, SQL, and BSON encoding support
- Comprehensive test suite for `Dyn` type covering all marshaling/unmarshaling interfaces
- Function documentation comments for all exported functions and methods

### Changed
- Improved error messages in `Nil` type with clearer expectations for encoding operations
- Refactored `Nil.Scan()` method to use switch statement for better readability
- Updated documentation for `Nil` and `Opt` types with cross-references
- Simplified inline error handling in unmarshal methods for better code consistency

## [v1.0.0] - 2025-06-28

### Added
- Initial release of typx package
- `Nil` type for representing nullable values with SQL, JSON, and BSON encoding support
- `Opt` type for representing optional values in JSON payloads
- Pointer helper functions for working with pointers
- GitHub Actions workflow for linting and testing
- Dependabot configuration for automated dependency updates
- Comprehensive test suite with high coverage
- MIT License
- Documentation and usage examples
