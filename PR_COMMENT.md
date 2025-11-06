# Fix: Modernize build tooling and add pre-commit hooks

## Background
The lack of consistent formatting in the main branch was leading to gratuitous differences between branches, making it difficult to review actual changes. This PR modernizes the build tooling and adds automatic formatting checks to prevent this issue going forward. The plan is to merge this to main first, which will unlock several bug fix and feature branches that can then be rebased on top of it.

## Summary
- Modernized Makefile with organized targets and better UX
- Updated golangci-lint config to v2 format with best practices
- Added tools.go for proper tool dependency management
- Updated developer documentation to match new tooling
- **Added pre-commit hooks for automatic code quality checks**
- **All existing lint issues have been fixed**

## Makefile Improvements
- Reorganized targets into logical categories (Development, Testing, Code Quality, etc.)
- Added comprehensive help system with target descriptions
- Implemented colored output for better visibility
- Added modern targets: coverage reports, benchmarks, dependency management
- Added safety checks for required tools
- Improved error messages and user feedback
- **Added `make install-hooks`, `make run-hooks`, and `make update-hooks` for pre-commit management**

## Linter Configuration (golangci-lint v2)
- Added version field required for v2 compatibility
- Fixed output format from array to map structure
- Removed deprecated linters (typecheck, gosimple, exportloopref)
- Replaced exportloopref with copyloopvar (new name in v2)
- Moved goimports and gofmt from linters to formatters
- Added comprehensive linter settings for better code quality
- Configured appropriate exclusions for tests and generated code
- Added limits on issue reporting to reduce noise
- **All lint issues resolved - project now passes `make lint` with 0 issues**

## Tool Management
- Created tools.go with build tag for tool dependencies
- Tools are now tracked in go.mod for version consistency
- Added installation instructions for tools that can't be vendored
- Documented golangci-lint v1.x compatibility issues with Go 1.23

## Pre-commit Hooks (New)
- **Added `.pre-commit-config.yaml` with Go-specific hooks**
- **Automatic checks on commit: go fmt, go vet, go mod tidy, golangci-lint, go build**
- **General code quality checks: trailing whitespace, file endings, YAML syntax, large files**
- **Auto-fixes common issues like whitespace and missing newlines**
- **Hooks run automatically after `make install-hooks` setup**

## Documentation Updates
- Updated CONTRIBUTING.md with new make targets
- Clarified tool installation process
- Added comprehensive testing and linting instructions
- Documented code generation workflow
- Organized commands by category for better discoverability
- **Added instructions for setting up git hooks with pre-commit**

## GitHub Workflows Updates
- Updated go.yml to use Makefile targets for consistency
- Standardized on Go 1.23.x (latest version only)
- Added coverage reporting job with Codecov integration
- Added go.mod verification job
- Updated golangci-lint.yml with version specification (v2.4)
- Added permissions and configuration options
- Created new ci.yml workflow for comprehensive CI checks
- Added format checking workflow to ensure code formatting

## Breaking Changes
None - all existing workflows continue to work

## Testing
All Makefile targets tested and working:
- ✅ make help - displays organized help
- ✅ make fmt - formats code
- ✅ make lint - runs linter (0 issues)
- ✅ make test - runs tests
- ✅ make build - builds binary
- ✅ make vet - runs go vet
- ✅ make tools - installs development tools
- **✅ make install-hooks - sets up pre-commit hooks**
- **✅ make run-hooks - all hooks passing**

## Notes
- golangci-lint is installed separately as a standalone binary (standard practice for Go projects)
- **Pre-commit automatically runs formatting and linting on every commit**
- **Tests excluded from pre-commit due to environment issues (run via `make test`)**
- S1005 exclusion applied only to specific test file (values/drop_test.go) where needed

## Checklist

- [ ] I have read the contribution guidelines.
- [ ] `make test` passes.
- [ ] `make lint` passes.
- [ ] New and changed code is covered by tests.
- [ ] Performance improvements include benchmarks.
- [ ] Changes match the *documented* (not just the *implemented*) behavior of Shopify.
