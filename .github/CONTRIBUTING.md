# Contributing to Raptor

First off, thank you for considering contributing to Raptor! It's people like you that make Raptor a great tool for everyone.

## Development Workflow

### Prerequisites

- **Go**: 1.26 or higher
- **Bun**: (Optional, for running the MCP inspector)
- **Docker**: (Required for running local tests)

### Setting up your environment

1. Fork and clone the repository.
2. Run the local test setup script:
   ```bash
   ./scripts/setup-test.sh
   source .env
   ```
3. Build the project:
   ```bash
   make build
   ```

### Running Tests

We value high-quality code. Before submitting a PR, please ensure all tests and linting pass:

```bash
# Run all tests
make test

# Run linter
make lint

# Run all checks (test + lint + vet)
make check
```

## Pull Request Process

1. Create a new branch for your feature or bugfix: `git checkout -b feature/my-new-feature`.
2. Commit your changes with clear, descriptive messages.
3. Ensure your code follows the existing style and is well-commented.
4. Push to your fork and submit a Pull Request.
5. Once your PR is submitted, our CI suite will run to ensure everything is correct.

## Code of Conduct

Please be respectful and professional in all interactions within the Raptor community.
