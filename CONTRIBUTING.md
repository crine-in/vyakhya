# Contributing to Vyākhyā

We are excited that you want to contribute to **Vyākhyā**! As an open-source project by **CRINE**, we welcome all kinds of contributions—from bug reports and documentation updates to performance optimizations and new feature implementations.

Please take a moment to review this document to ensure a smooth contribution process.

---

## 📜 Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please report any unacceptable behavior to [support@crine.in](mailto:support@crine.in).

---

## 🛠️ How to Contribute

### 1. Reporting Bugs
If you find a bug:
- Check if the bug is already reported in the Github Issues section.
- If not, open a new issue.
- Describe the bug clearly: what did you expect, what actually happened, and steps to reproduce.
- Include logs, system information (OS/Go version), and sample API payloads if applicable.

### 2. Suggesting Enhancements
We are always looking to make Vyākhyā faster and more feature-rich:
- Open a feature request issue.
- Describe the proposed feature, the use case, and any architectural ideas you might have.

### 3. Submitting Pull Requests
1. **Fork the Repository**: Create a fork of the official repository.
2. **Clone Locally**: Clone your fork to your development machine.
3. **Create a Branch**: Create a descriptive feature branch (e.g., `git checkout -b feature/optimize-unmarshaler`).
4. **Write Code**: Implement your changes. Make sure you follow our [Coding Standards](#-coding-standards).
5. **Add Tests**: Write unit tests to cover your new logic. Ensure all tests and benchmarks pass:
   ```bash
   go test ./... -v
   go test -bench=. ./wordnet
   ```
6. **Commit Changes**: Write clear, descriptive commit messages.
7. **Push & PR**: Push the branch to your fork and submit a Pull Request to the `main` branch of the official repository.

---

## 🎨 Coding Standards

To maintain Vyākhyā's peak speed and minimal memory footprint, please adhere to the following standards:

1.  **Standard Library Only**: We enforce a **zero-dependency** design. Do not add external Go packages (no third-party routers, JSON parsers, etc.). All features must be implemented using Go's native standard library packages.
2.  **Performance Mindset**: Ensure that your lookup paths are highly optimized. Avoid unnecessary allocations in the search hot path. Run benchmark tests before and after your changes:
    ```bash
    go test -bench=. ./wordnet -benchmem
    ```
3.  **Formatting**: Go code must be formatted using the standard tool:
    ```bash
    go fmt ./...
    ```
4.  **Testing**: All code additions must be accompanied by comprehensive tests in the appropriate package.
5.  **Licensing**: Prepend the standard AGPL-3.0 copyright header to any new `.go` files:
    ```go
    // Copyright (C) 2026 CRINE (https://www.crine.in) <support@crine.in>
    //
    // This program is free software: you can redistribute it and/or modify ...
    ```

---

## 📥 Pull Request Checklist

Before submitting your PR, please verify:
- [ ] Code is formatted with `go fmt`.
- [ ] All package unit tests pass.
- [ ] Benchmark performance has not regressed.
- [ ] Copyright header is present at the top of new files.
- [ ] Changes are documented in the code or README if applicable.

---

## 📧 Need Help?

If you have any questions or need architectural guidance, feel free to contact us:
- **Website**: [www.crine.in](https://www.crine.in)
- **General contact**: [contact@crine.in](mailto:contact@crine.in)
- **Support**: [support@crine.in](mailto:support@crine.in)
