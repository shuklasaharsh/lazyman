# Contributors Guide

Thank you for your interest in contributing to LazyMan!

## Getting Started

LazyMan is a hobby project maintained by a single developer in their spare time. Your patience with PR reviews is greatly appreciated.

## Contribution Workflow

### 1. Create an Issue

Before starting work on a contribution:

1. **Search existing issues** to avoid duplicates
2. **Create a new issue** describing:
   - The problem you want to solve or feature you want to add
   - Your proposed approach or solution
   - Any alternatives you've considered
   - Screenshots or examples if applicable
3. **Wait for issue acceptance** and discussion
4. Once accepted, the issue may be assigned to you or marked as ready for implementation

### 2. Development

Once your issue is accepted:

1. Fork the repository
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/lazymanuals.git
   cd lazymanuals
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. Make your changes:
   - Follow existing code style and conventions
   - Write clear, descriptive commit messages
   - Keep commits focused and logical
5. Build and test:
   ```bash
   go build
   # Test the application manually
   ```

### 3. Submit a Pull Request

1. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
2. Create a PR against the `main` branch
3. In your PR description, include:
   - Reference to the issue: "Closes #123" or "Fixes #123"
   - Summary of changes made
   - How you tested the changes
   - Screenshots/recordings if UI changes
   - Any breaking changes or migration notes

### 4. Code Review

- The maintainer will review when time permits
- Be responsive to feedback and questions
- Make requested changes in new commits
- Once approved, your PR will be merged

## Development Guidelines

### Code Style
- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Keep functions focused and reasonably sized
- Add comments for non-obvious logic

### Testing
- Manually test your changes thoroughly
- Verify the application builds without errors
- Test edge cases and error conditions

### Commit Messages
- Use clear, descriptive commit messages
- Start with a verb in present tense: "Add", "Fix", "Update", "Remove"
- Keep the first line under 72 characters
- Add details in the body if needed

### General Guidelines
- Keep PRs focused on a single issue or feature
- Avoid unrelated refactoring or style changes
- Don't include personal IDE or system files
- Update documentation if your changes affect usage

## Questions or Problems?

If you have questions or run into issues:
- Check existing issues and discussions
- Create a new issue with your question
- Be patient - responses may take time

## Code of Conduct

- Be respectful and constructive
- Help create a welcoming environment
- Focus on the code and technical discussion
- Remember this is a hobby project - be kind!

## Recognition

All contributors will be acknowledged for their contributions to the project.

Thank you for contributing to LazyMan! 🎉
