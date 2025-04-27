# Contributing to TaskQueue

Thank you for your interest in contributing to TaskQueue! We welcome contributions of all kinds, whether it's reporting bugs, suggesting new features, improving documentation, or writing code.

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

There are many ways to contribute to TaskQueue:

### Reporting Issues

If you encounter a bug or unexpected behavior, please help us by creating a clear and detailed issue. When reporting a bug, please include the following information:

* **A clear and concise description of the bug.**
* **Steps to reproduce the behavior.**
* **Expected behavior.**
* **Go version (`go version` output).**
* **TaskQueue version.**
* **Operating System.**
* **Any relevant error messages or logs.**

Please check the existing issues to avoid submitting duplicates.

### Suggesting Features and Enhancements

We are always open to suggestions for new features and improvements. When proposing a feature, please consider:

* **Clearly describe the proposed feature or enhancement.**
* **Explain the problem it solves or the benefit it provides.**
* **Discuss any potential alternatives you've considered.**
* **Provide any relevant context or use cases.**

Feel free to open an issue with your feature request for discussion.

### Contributing Code

If you'd like to contribute code to TaskQueue, please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your changes. Please follow our [Branch Naming Convention](#branch-naming-convention).
3. Follow the project's coding style. While we don't have a strict style guide documented yet, please aim for clean, readable, and idiomatic Go code. Pay attention to naming conventions, formatting (using `gofmt` is recommended), and clear comments.
4. Write tests for your changes. New features should have accompanying tests to ensure they function correctly, and bug fixes should include tests that reproduce the bug and verify the fix.
5. Ensure all tests pass before submitting a pull request. You can run tests using `go test ./...`.
6. Document your code. Add clear and concise doc comments to any new functions, methods, types, or packages you introduce.
7. Make small, focused pull requests. It's easier to review and merge smaller PRs.
8. Clearly describe your changes in the pull request description. Explain the purpose of your changes and any relevant context. Please also adhere to our [Commit Message Convention](#commit-message-convention).
9. Link any related issues in your pull request description (e.g., "Fixes #123").

### Improving Documentation

Contributions to the documentation are highly valued. You can help by:

* **Fixing typos and grammatical errors.**
* **Clarifying existing documentation.**
* **Adding documentation for new features.**
* **Creating examples and tutorials.**

Documentation is typically written in Markdown. Please submit documentation changes via pull requests.

## Getting Started with Development

1. **Install Go:** Make sure you have a recent version of Go installed on your system. You can download it from [https://go.dev/dl/](https://go.dev/dl/).
2. **Clone the repository:**

    ```bash
    git clone https://github.com/KengoWada/taskqueue.git
    cd taskqueue
    ```

3. **Explore the codebase:** Familiarize yourself with the project structure and existing code.
4. **Start contributing!**

## Branch Naming Convention

We use the following branch naming convention:

`[<issue-number>-]<type>/<simple-description>`

* **`<issue-number>` (optional):** The number of the issue this branch addresses (e.g., `42`). Include this if the branch is specifically for an issue.
* Where `<type>` is one of:

  * `feature`: For new features or enhancements.
  * `fix`: For bug fixes.
  * `docs`: For documentation changes.
  * `refactor`: For code refactoring.
  * `test`: For changes related to tests.
  * `chore`: For maintenance tasks.

`<short-description>` is a concise, lowercase description using hyphens as word separators.

**Examples:**

* **For an issue:** `42-feature/add-redis-broker`
* **For an issue:** `57-fix/handle-redis-broker-error`
* **For an issue:** `63-docs/improve-contribution-section`
* **General refactoring (no specific issue):** `refactor/simplify-error-handling`
* **Adding general tests:** `test/add-integration-tests`
* **Updating dependencies (no specific issue):** `chore/update-go-dependencies`

When creating a branch to work on a specific issue, please prefix your branch name with the issue number followed by a hyphen and the type of change. If the branch is for a general improvement or task not tied to a specific issue, you can omit the issue number prefix.

## Commit Message Convention

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for our commit messages. This provides a structured format that helps automate changelog generation and understand the purpose of commits.

Commit messages should be structured as follows:

```txt
<type>(<scope>): <short description>

[optional body]

[optional footer(s)]
```

* **`<type>`:** Indicates the category of the commit (e.g., `feat`, `fix`, `docs`, `refactor`, `test`, `build`, `ci`, `chore`, `revert`).
* **`<scope>` (optional):** Specifies the part of the codebase affected (e.g., `broker`, `worker`, `manager`).
* **`<short description>`:** A concise summary in the imperative, present tense (e.g., "Add Redis broker implementation"). Capitalize the first word and do not end with a period.
* **`[optional body]`:** A more detailed explanation of the changes, separated from the subject by a blank line. Explain the *what*, *why*, and *how*.
* **`[optional footer(s)]`:** Can include issue tracking references (`Fixes #123`, `Closes #456`), breaking change notes (`BREAKING CHANGE: ...`), and co-authored-by information.

**Examples:**

```txt
feat(broker): Add Redis broker implementation

- Implemented RedisBroker to enable task queuing using Redis.
- Created Publish and Consume methods for interacting with Redis queues.
- Added dependency to go-redis package.

Closes #15
```

```txt
feat(broker): Implement Redis broker for task queuing

- Added RedisBroker with Publish and Consume methods for task handling.
- Integrated go-redis client for Redis communication.
- Updated example to demonstrate RedisBroker usage.
```

## Pull Request Process

1. Once you've made your changes and your commits follow the [Commit Message Convention](#commit-message-convention), push your branch to your forked repository.
2. Create a new pull request (PR) on the main TaskQueue repository.
3. Ensure your PR description clearly explains the changes and links any related issues.
4. Maintainers will review your PR. Be prepared to address any feedback or make necessary revisions.
5. Once your PR is approved and all checks pass, it will be merged into the main branch.

Thank you again for your contribution! We appreciate your help in making TaskQueue a better project.
