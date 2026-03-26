# Contributing to simplelogin-cli

## Building from source

Requirements: Go 1.24+

```bash
git clone https://github.com/mexcool/simplelogin-cli.git
cd simplelogin-cli
make build
# Binary output: ./bin/sl
```

You can also install directly to `/usr/local/bin`:

```bash
make install
```

Or use `go build` directly:

```bash
go build -o bin/sl ./cmd/sl
```

## Running tests

```bash
make test
# or
go test ./...
```

## Code style

- Format all code with `gofmt` before committing.
- Run `go vet ./...` and fix any issues before opening a PR.
- Keep changes focused — one logical change per PR.

## Submitting a PR

1. Fork the repository and create a branch from `main`:
   ```bash
   git checkout -b your-feature-or-fix
   ```
2. Make your changes, commit them, and push to your fork.
3. Open a pull request against `main` with a clear title and description of what changed and why.
4. For bug fixes, include steps to reproduce the original issue.

## Reporting issues

Use the [GitHub issue tracker](https://github.com/mexcool/simplelogin-cli/issues). Include your OS, Go version, and the exact command and output that demonstrates the problem.

## Security vulnerabilities

Please do **not** open a public issue for security vulnerabilities. Email **cli.chewy305@simplelogin.com** instead. See [SECURITY.md](SECURITY.md) for details.

## Contact

For questions or help, reach out at **cli.chewy305@simplelogin.com**.
