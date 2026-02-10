# AI Agent Contribution Guide

This document provides guidelines for AI agents contributing to LazyMan.

## Important Context

LazyMan is a **hobby project** maintained by a single developer. Please be patient with PR reviews - they may take some time due to limited availability.

## Recommended Workflow

### 1. Create an Issue First (Strongly Recommended)

Before creating a PR, please:
1. **Create an issue** describing the proposed change
2. Include:
   - Clear description of the problem or enhancement
   - Proposed solution or approach
   - Any relevant context or examples
3. **Wait for issue acceptance** before proceeding with implementation
4. Reference the issue number in your PR

This approach:
- Saves time by validating the change before implementation
- Allows discussion of design decisions
- Ensures alignment with project goals
- Reduces unnecessary PRs

### 2. Fork and Submit PR

If you choose to proceed directly or after issue acceptance:

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes:
   - Follow existing code style and conventions
   - Ensure code builds successfully (`go build`)
   - Test your changes thoroughly
4. Commit with clear, descriptive messages
5. Submit a PR to the `main` branch with:
   - Clear title and description
   - Reference to related issue (if exists)
   - Summary of changes made
   - Any testing performed

## Code Guidelines

- Follow Go best practices and idioms
- Match existing code formatting
- Keep changes focused and minimal
- Avoid unnecessary refactoring in feature PRs
- Ensure no breaking changes without discussion

## Review Timeline

As this is a hobby project with one developer:
- Reviews may take several days or longer
- Be patient and respectful
- The maintainer will respond when time permits
- Issues help prioritize work

## Questions?

If you have questions about contributing, please create an issue for discussion.
