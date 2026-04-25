# Contributing to worklog

Thank you for taking the time to contribute!

## Reporting issues

- Search [existing issues](https://github.com/jquiaios/worklog/issues) before opening a new one.
- Include your OS, Go version, and the exact command or steps to reproduce.

## Proposing changes

For anything non-trivial (new features, behaviour changes, large refactors), **open an issue first** to discuss the approach before writing code. This avoids wasted effort if the direction doesn't fit the project.

Small fixes (typos, obvious bugs, documentation) can go straight to a pull request.

## Development setup

The project uses Docker so you don't need Go installed locally:

```
git clone https://github.com/jquiaios/worklog.git
cd worklog
docker build -t worklog .
mkdir -p ~/.worklog
docker run -it --rm --user $(id -u):$(id -g) -v ~/.worklog:/home/wl/.worklog worklog
```

If you prefer a local Go setup, Go 1.25+ is required.

## Submitting a pull request

1. Fork the repo and create a branch from `main`.
2. Make your changes and ensure existing tests pass (`go test ./...`).
3. Add tests for new behaviour where it makes sense.
4. Keep commits focused — one logical change per commit.
5. Open a pull request against `main` with a clear description of what and why.

## Code style

- Run `go vet ./...` and `go fmt ./...` before pushing.
- Follow standard Go conventions — no framework-specific style rules.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](./LICENSE).
