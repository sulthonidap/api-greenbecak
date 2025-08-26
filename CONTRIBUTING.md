# Contributing to GreenBecak Backend

## Overview

Terima kasih atas minat Anda untuk berkontribusi pada GreenBecak Backend! Guide ini akan membantu Anda memahami cara berkontribusi dengan efektif.

## Getting Started

### 1. Prerequisites

- Go 1.21+
- MySQL 8.0+
- Git
- Docker (optional)

### 2. Fork and Clone

```bash
# Fork repository di GitHub
# Clone repository Anda
git clone https://github.com/YOUR_USERNAME/greenbecak-backend.git
cd greenbecak-backend

# Add upstream remote
git remote add upstream https://github.com/greenbecak/backend.git
```

### 3. Setup Development Environment

```bash
# Install dependencies
go mod tidy

# Copy environment file
cp env.example .env

# Edit environment variables
nano .env

# Setup database
mysql -u root -p
CREATE DATABASE greenbecak_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. Run Application

```bash
# Development mode
go run main.go

# Or using Makefile
make run
```

## Development Workflow

### 1. Create Feature Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/your-bug-description
```

### 2. Make Changes

- Write clean, readable code
- Follow Go coding conventions
- Add tests for new features
- Update documentation

### 3. Commit Changes

```bash
# Add changes
git add .

# Commit with conventional message
git commit -m "feat: add user authentication endpoint"

# Push to your fork
git push origin feature/your-feature-name
```

### 4. Create Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select your feature branch
4. Fill out the PR template
5. Submit PR

## Code Standards

### 1. Go Conventions

```go
// Use camelCase for variables and functions
func getUserByID(id uint) (*User, error) {
    // Implementation
}

// Use PascalCase for exported functions
func GetUserByID(id uint) (*User, error) {
    // Implementation
}

// Use meaningful names
var userCount int // Good
var c int         // Bad
```

### 2. File Organization

```
backend/
├── cmd/           # Command line tools
├── config/        # Configuration
├── database/      # Database operations
├── handlers/      # HTTP handlers
├── middleware/    # Middleware
├── models/        # Data models
├── routes/        # Route definitions
├── utils/         # Utility functions
└── docs/          # Documentation
```

### 3. Error Handling

```go
// Always check errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Use custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
```

### 4. Testing

```go
// Write tests for all new code
func TestCreateUser(t *testing.T) {
    // Arrange
    req := CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    // Act
    user, err := CreateUser(req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, req.Username, user.Username)
}
```

## Pull Request Guidelines

### 1. PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes
```

### 2. Review Process

1. **Self Review**: Review your own code before submitting
2. **Tests**: Ensure all tests pass
3. **Documentation**: Update relevant documentation
4. **Squash Commits**: Clean up commit history if needed

### 3. Code Review Checklist

- [ ] Code is readable and well-documented
- [ ] Tests are comprehensive
- [ ] Error handling is appropriate
- [ ] Security considerations addressed
- [ ] Performance impact considered
- [ ] No breaking changes (unless intentional)

## Issue Guidelines

### 1. Bug Reports

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: Ubuntu 20.04
- Go version: 1.21.0
- Database: MySQL 8.0

## Additional Information
Screenshots, logs, etc.
```

### 2. Feature Requests

```markdown
## Feature Description
Clear description of the feature

## Use Case
Why this feature is needed

## Proposed Solution
How you think it should be implemented

## Alternatives Considered
Other approaches you considered

## Additional Information
Any other relevant information
```

## Development Tools

### 1. Code Formatting

```bash
# Format code
go fmt ./...

# Or using Makefile
make fmt
```

### 2. Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Or using Makefile
make lint
```

### 3. Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Or using Makefile
make test
```

### 4. Pre-commit Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running pre-commit checks..."

# Format code
go fmt ./...

# Run tests
go test ./...

# Run linter
golangci-lint run

echo "Pre-commit checks completed!"
```

Make executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Documentation

### 1. Code Comments

```go
// CreateUser creates a new user in the database
// Returns the created user or an error if creation fails
func CreateUser(req CreateUserRequest) (*User, error) {
    // Implementation
}
```

### 2. API Documentation

```go
// @Summary Create new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /api/users [post]
func CreateUser(c *gin.Context) {
    // Implementation
}
```

### 3. README Updates

- Update README.md for new features
- Add examples for new endpoints
- Update installation instructions
- Add troubleshooting section if needed

## Security Guidelines

### 1. Input Validation

```go
// Always validate input
func validateEmail(email string) error {
    if email == "" {
        return errors.New("email is required")
    }
    
    if !strings.Contains(email, "@") {
        return errors.New("invalid email format")
    }
    
    return nil
}
```

### 2. SQL Injection Prevention

```go
// Use parameterized queries
db.Where("email = ?", email).First(&user)

// Don't use string concatenation
// db.Where("email = '" + email + "'").First(&user) // BAD
```

### 3. Authentication & Authorization

```go
// Always check permissions
func requireAdmin(c *gin.Context) {
    role, exists := c.Get("role")
    if !exists || role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
        c.Abort()
        return
    }
}
```

## Performance Guidelines

### 1. Database Optimization

```go
// Use indexes for frequently queried fields
// Add database indexes in migrations

// Use pagination for large datasets
func GetUsers(page, limit int) ([]User, error) {
    offset := (page - 1) * limit
    return db.Offset(offset).Limit(limit).Find(&users)
}
```

### 2. Caching

```go
// Implement caching for expensive operations
var userCache = make(map[uint]*User)

func GetUserByID(id uint) (*User, error) {
    if user, exists := userCache[id]; exists {
        return user, nil
    }
    
    // Fetch from database
    user := &User{}
    if err := db.First(user, id).Error; err != nil {
        return nil, err
    }
    
    // Cache result
    userCache[id] = user
    return user, nil
}
```

### 3. Connection Pooling

```go
// Configure connection pool
sqlDB, err := db.DB()
if err != nil {
    return err
}

sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

## Troubleshooting

### 1. Common Issues

**Tests Failing**
```bash
# Check database connection
mysql -u test_user -p test_database

# Run tests with verbose output
go test -v ./...
```

**Build Errors**
```bash
# Clean and rebuild
go clean
go mod tidy
go build
```

**Linting Errors**
```bash
# Fix auto-fixable issues
golangci-lint run --fix

# Check specific rules
golangci-lint run --disable-all --enable=govet
```

### 2. Getting Help

- **GitHub Issues**: Create an issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Slack**: Join our Slack channel for real-time help
- **Email**: Contact maintainers directly

## Release Process

### 1. Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### 2. Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version bumped
- [ ] Release notes written
- [ ] Tagged and released

### 3. Creating a Release

```bash
# Update version
git tag v1.0.0

# Push tag
git push origin v1.0.0

# Create release on GitHub
# Add release notes and assets
```

## Community Guidelines

### 1. Be Respectful

- Be kind and respectful to others
- Use inclusive language
- Welcome newcomers
- Give constructive feedback

### 2. Communication

- Use clear, concise language
- Provide context for issues
- Be patient with responses
- Follow up on discussions

### 3. Recognition

- Contributors will be credited in releases
- Significant contributions will be highlighted
- Long-term contributors may become maintainers

## Getting Help

### 1. Resources

- **Documentation**: https://docs.greenbecak.com
- **API Reference**: https://docs.greenbecak.com/api
- **Examples**: https://github.com/greenbecak/examples
- **Blog**: https://blog.greenbecak.com

### 2. Contact

- **GitHub Issues**: https://github.com/greenbecak/backend/issues
- **Discussions**: https://github.com/greenbecak/backend/discussions
- **Email**: contributors@greenbecak.com
- **Slack**: #greenbecak-contributors

### 3. Mentorship

- New contributors can request mentorship
- Experienced contributors can offer to mentor
- Pair programming sessions available
- Code review guidance provided

## Recognition

### 1. Contributors

All contributors will be listed in:
- README.md contributors section
- Release notes
- Project website

### 2. Special Recognition

- **Core Contributors**: Long-term, significant contributions
- **Bug Hunters**: Finding and fixing critical bugs
- **Documentation Heroes**: Improving documentation
- **Community Builders**: Helping others contribute

### 3. Hall of Fame

- Contributors with 10+ PRs
- Contributors with 100+ commits
- Contributors with major feature implementations
- Contributors with security improvements

Thank you for contributing to GreenBecak Backend! 🚀
