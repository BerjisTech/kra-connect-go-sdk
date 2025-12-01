# Contributing to KRA-Connect Go SDK

Thank you for your interest in contributing to the KRA-Connect Go SDK! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and constructive in your interactions with other contributors.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-sdk.git
   cd go-sdk
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/BerjisTech/kra-connect-go-sdk.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for running make commands)

### Install Dependencies

```bash
go mod download
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running Examples

```bash
# Set your API key
export KRA_API_KEY="your-api-key-here"

# Run basic example
go run examples/basic/main.go

# Run batch example
go run examples/batch/main.go

# Run error handling example
go run examples/error-handling/main.go

# Run advanced example
go run examples/advanced/main.go
```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/add-webhook-support`
- `fix/authentication-timeout`
- `docs/update-readme`
- `refactor/simplify-error-handling`

### Commit Messages

Follow the conventional commits specification:

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(client): add webhook support for real-time updates

fix(auth): handle token refresh race condition

docs(readme): update installation instructions

test(validator): add tests for edge cases
```

## Testing

### Test Coverage

- Maintain at least 80% test coverage
- Write tests for all new features
- Update tests when modifying existing code

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Code Style

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

### Key Principles

1. **Use `gofmt`**: Format all code with `gofmt`
   ```bash
   gofmt -w .
   ```

2. **Use `go vet`**: Check for common mistakes
   ```bash
   go vet ./...
   ```

3. **Use `golint`**: Check for style issues
   ```bash
   golint ./...
   ```

4. **Documentation**: All exported functions, types, and constants must have GoDoc comments
   ```go
   // Client is the main KRA Connect client
   //
   // The client is goroutine-safe and can be used concurrently.
   //
   // Example:
   //
   //  client, err := kra.NewClient(
   //      kra.WithAPIKey("your-api-key"),
   //  )
   type Client struct {
       // ...
   }
   ```

5. **Error Handling**: Always handle errors explicitly
   ```go
   // Good
   result, err := client.VerifyPIN(ctx, pin)
   if err != nil {
       return nil, fmt.Errorf("failed to verify PIN: %w", err)
   }

   // Bad
   result, _ := client.VerifyPIN(ctx, pin)
   ```

6. **Context**: Use context for cancellation and timeouts
   ```go
   func (c *Client) VerifyPIN(ctx context.Context, pin string) (*PINVerificationResult, error) {
       // Implementation
   }
   ```

7. **Naming Conventions**:
   - Use `camelCase` for unexported names
   - Use `PascalCase` for exported names
   - Use meaningful names (avoid single-letter variables except in short scopes)
   - Use `ID` not `Id` (same for `URL`, `HTTP`, etc.)

## Submitting Changes

### Pull Request Process

1. Update your fork with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Ensure all tests pass:
   ```bash
   go test ./...
   ```

3. Update documentation if needed

4. Commit your changes with clear commit messages

5. Push to your fork:
   ```bash
   git push origin your-branch-name
   ```

6. Create a pull request on GitHub

### Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Include examples if adding new features
- Ensure CI checks pass
- Respond to review comments promptly

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe the tests you ran and their results

## Checklist
- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
```

## Reporting Bugs

### Before Submitting a Bug Report

1. Check the [issue tracker](https://github.com/BerjisTech/kra-connect-go-sdk/issues) for existing reports
2. Ensure you're using the latest version
3. Verify the bug is reproducible

### Bug Report Template

```markdown
**Describe the bug**
A clear description of the bug

**To Reproduce**
Steps to reproduce the behavior:
1. Create client with '...'
2. Call method '...'
3. See error

**Expected behavior**
What you expected to happen

**Code example**
```go
// Minimal code example that demonstrates the bug
```

**Environment**
- Go version: [e.g., 1.21.0]
- SDK version: [e.g., 0.1.3]
- OS: [e.g., Ubuntu 22.04]

**Additional context**
Any other relevant information
```

## Feature Requests

We welcome feature requests! Please:

1. Check existing feature requests first
2. Provide a clear use case
3. Explain why the feature would be valuable
4. Consider submitting a pull request if you can implement it

### Feature Request Template

```markdown
**Is your feature request related to a problem?**
A clear description of the problem

**Describe the solution you'd like**
What you want to happen

**Describe alternatives you've considered**
Other solutions you've thought about

**Additional context**
Any other relevant information

**Example usage**
```go
// Show how you'd like to use the feature
```
```

## Development Workflow

### Adding a New Feature

1. Create a feature branch
2. Implement the feature with tests
3. Update documentation
4. Submit a pull request

### Fixing a Bug

1. Create a bug fix branch
2. Write a test that reproduces the bug
3. Fix the bug
4. Ensure all tests pass
5. Submit a pull request

## Questions?

If you have questions about contributing, please:

1. Check the [documentation](https://pkg.go.dev/github.com/BerjisTech/kra-connect-go-sdk)
2. Search [existing issues](https://github.com/BerjisTech/kra-connect-go-sdk/issues)
3. Create a new issue with the `question` label

Thank you for contributing to KRA-Connect Go SDK!
