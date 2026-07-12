# Python Packaging & Tooling

Modern Python is one declarative file (`pyproject.toml`) plus fast tooling (`uv`, `ruff`, `pyright`).
Avoid `setup.py`, `requirements.txt` sprawl, and global `pip install`.

## pyproject.toml (PEP 621) — single source of truth

```toml
[project]
name = "myapp"
version = "0.1.0"
requires-python = ">=3.12"
dependencies = [
    "httpx>=0.27",
    "pydantic>=2.7",
]

[project.optional-dependencies]
dev = ["pytest>=8", "pytest-cov", "ruff", "pyright"]

[project.scripts]
myapp = "myapp.cli:main"        # console entry point

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

## uv — env + dependency management (fast)

`uv` replaces `pip` + `venv` + `pip-tools` and is dramatically faster. It writes a `uv.lock` for
reproducible installs.

```
uv venv                       # create .venv
uv add httpx                  # add a dep (updates pyproject + lock)
uv add --dev pytest ruff      # dev dep
uv sync                       # install exactly what the lock pins
uv run pytest                 # run inside the managed env
```

Plain pip fallback (no uv): `python -m venv .venv && . .venv/bin/activate && pip install -e ".[dev]"`.

## src layout — use it

```
myapp/
  pyproject.toml
  src/
    myapp/
      __init__.py
      cli.py
  tests/
    test_cli.py
```

`src/` prevents importing the package from the working tree instead of the installed copy — tests then
exercise what users actually install. Install editable for development:

```
uv pip install -e .          # or: pip install -e .
```

## ruff — lint + format (replaces black, isort, flake8)

```toml
[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "UP", "B", "SIM"]   # errors, pyflakes, imports, pyupgrade, bugbear, simplify
```

```
ruff check --fix        # lint + autofix
ruff format             # format (black-compatible)
```

## pyright / mypy config

```toml
[tool.pyright]
typeCheckingMode = "strict"
pythonVersion = "3.12"
```

or mypy:

```toml
[tool.mypy]
python_version = "3.12"
strict = true
warn_unreachable = true
```

Run type checking in CI — it's the highest-leverage gate for a dynamic language.

## Configuration & secrets

- Read config from environment (`os.environ`) or a settings model (`pydantic-settings`); never hardcode.
- Keep secrets out of the repo and out of `pyproject.toml`; load from env/secret store.
- `logging`, not `print`, for runtime diagnostics.

## Rules

- One `pyproject.toml`; no `setup.py`/`requirements.txt` unless a legacy project forces it.
- Lock dependencies (`uv.lock`) for reproducible installs.
- `src/` layout so tests run against the installed package.
- `ruff` for lint+format, `pyright`/`mypy --strict` for types — both in CI.
- Pin `requires-python`; target a concrete version in ruff/pyright config.
