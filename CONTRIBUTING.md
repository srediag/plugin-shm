# How to Contribute

Welcome to **plugin-shm**!  
We use GitHub for our codebase and contributions. This document will guide you on how to contribute effectively.

---

## Your First Pull Request

You can start by reading [How To Pull Request](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests).

---

## Branch Organization

We use a simple branching model inspired by [git-flow](https://nvie.com/posts/a-successful-git-branching-model/) and [Feature-Driven Development (FDD)](https://en.wikipedia.org/wiki/Feature-driven_development).

- `main`: Always stable, ready for release.
- `develop`: Active development happens here.
- `feature/*`: Feature-specific branches merged into `develop`.
- `hotfix/*`: Critical patches merged into `main` and `develop`.

---

## Bugs

### 1. How to Find Known Issues

We track public issues via [GitHub Issues](https://github.com/srediag/plugin-shm/issues).  
Before reporting a new issue, **please search** to avoid duplications.

### 2. Reporting New Issues

Please provide a **reduced reproducible example** whenever possible.  
Examples can be:

- Embedded directly in the GitHub issue.
- A link to the [Golang Playground](https://play.golang.org/).

### 3. Security Bugs

For security vulnerabilities, **please do not open a public issue**.  
Instead, report them via email: [plugin-shm security contact](mailto:security@srediag.io).

---

## How to Get in Touch

- Open an [Issue](https://github.com/srediag/plugin-shm/issues) for bugs or discussions.
- Contact our maintainers at [security@srediag.io](mailto:security@srediag.io) for security-related matters.

---

## Submit a Pull Request

Before submitting a Pull Request (PR), consider the following:

1. Search [GitHub Pull Requests](https://github.com/srediag/plugin-shm/pulls) for open or closed PRs related to your change.
2. If proposing a significant design or refactor, **open an Issue first** to discuss.
3. Fork the repository:

    ```bash
    git clone https://github.com/srediag/plugin-shm.git
    cd plugin-shm
    git checkout -b my-feature-branch develop
    ```
