# Contributing to GDF

Thank you for your interest in contributing to GDF (Go Dotfiles)! We welcome contributions from the community to help make this project better.

## Getting Started

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally:
    ```bash
    git clone https://github.com/YOUR_USERNAME/GoDotFiles.git
    cd GoDotFiles
    ```
3.  **Install Go**: Ensure you have Go 1.21+ installed.

## Development Workflow

1.  **Create a branch** for your feature or bugfix:
    ```bash
    git checkout -b feature/my-new-feature
    ```
2.  **Write code and tests**. We follow Test-Driven Development (TDD). 
    -   Run tests: `make test`
    -   Run linting: `make lint`
3.  **Update Documentation**: If you change behavior, update the relevant docs in `docs/`.
4.  **Commit changes** using [Conventional Commits](https://www.conventionalcommits.org/):
    ```bash
    git commit -m "feat(cli): add new command"
    ```

## Release Process

For maintainers, please refer to the [Release Guide](docs/development/release.md) for instructions on how to cut a new release.
