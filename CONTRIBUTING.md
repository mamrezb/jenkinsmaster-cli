
# Contributing to JenkinsMaster CLI

First off, thank you for considering contributing to JenkinsMaster CLI! Your contributions help make this project better and more useful for the community.

## Table of Contents

- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
  - [Bug Reports](#bug-reports)
  - [Feature Requests](#feature-requests)
  - [Code Contributions](#code-contributions)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Community Guidelines](#community-guidelines)

---

## Getting Started

To contribute to this project, ensure that you:

1. Have read the [README.md](README.md) file to understand the project and its scope.
2. Have a GitHub account to fork and clone the repository.
3. Are familiar with Go, Terraform, and Ansible, which are core technologies used in this project.

---

## How to Contribute

### Bug Reports

If you find a bug, please open an issue in the [GitHub Issues](https://github.com/mamrezb/jenkinsmaster-cli/issues) section with the following details:

1. A clear and descriptive title.
2. Steps to reproduce the issue.
3. Your environment details (OS, version, etc.).
4. Logs or screenshots to demonstrate the issue.

### Feature Requests

Have a feature in mind? Open a feature request issue with:

1. A clear and concise description of the feature.
2. The problem it solves or the improvement it provides.
3. Any relevant examples or prior art.

### Code Contributions

If you're ready to contribute code:

1. Fork the repository on GitHub.
2. Clone your fork locally.
3. Create a new branch for your changes: `git checkout -b feature/your-feature`.
4. Write clear and well-documented code.
5. Add tests to validate your changes.
6. Push your changes and open a Pull Request (PR).

---

## Code Style

Follow these guidelines to maintain consistency:

1. **Go Code**:
   - Use `go fmt` to format your code.
   - Follow idiomatic Go practices.

2. **Terraform Modules**:
   - Use clear and descriptive variable names.
   - Ensure all variables are documented in the `README.md` for the module.

3. **Ansible Playbooks**:
   - Follow YAML best practices.
   - Use roles and modules appropriately to keep playbooks modular and reusable.

---

## Testing

Testing is critical to ensure stability and maintainability. Before submitting a pull request:

1. Run all existing tests to ensure your changes don’t break anything:
   ```bash
   go test ./...
   ```
2. If adding a new feature or fixing a bug, include tests that validate the behavior.
3. Verify integration with Terraform and Ansible where applicable.

---

## Pull Request Process

To ensure a smooth PR review process:

1. Ensure your PR title is descriptive and references any related issues (e.g., "Fix: Address nil pointer exception (#42)").
2. Provide a clear description of what the PR does, including screenshots or logs if applicable.
3. Make sure your branch is up-to-date with the main branch: 
   ```bash
   git fetch origin
   git rebase origin/main
   ```
4. Address any feedback promptly during the review process.

---

## Community Guidelines

- Be respectful and professional in all communications.
- Provide constructive feedback when reviewing code.
- Collaborate openly and transparently.

We look forward to your contributions!

Thank you,  
The JenkinsMaster CLI Team