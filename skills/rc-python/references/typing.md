# Python Typing & Generics (3.12+)

Precise types are the cheapest correctness tool in Python. Target `pyright` (or `mypy --strict`) with
zero errors. Type every public signature; internal helpers may infer.

## Built-in generics and unions

Use built-in collection generics and `|` unions — no `typing.List`/`Optional` needed on 3.10+.

```python
def totals(rows: list[dict[str, int]]) -> dict[str, int]: ...

def find(name: str) -> User | None: ...   # not Optional[User]
```

`from __future__ import annotations` makes all annotations lazy strings (cheap, avoids forward-reference
quoting). Safe default for library code; note it changes runtime introspection (e.g. some Pydantic/
dataclass edge cases) — omit it when a framework reads annotations at runtime.

## PEP 695 generics (3.12+)

Native syntax — no `TypeVar` boilerplate.

```python
def first[T](items: list[T]) -> T | None:
    return items[0] if items else None

class Box[T]:
    def __init__(self, value: T) -> None:
        self.value = value

type Vector = list[float]          # type alias statement
type Result[T] = T | Exception     # generic alias
```

Bounds and constraints:

```python
def largest[T: (int, float)](xs: list[T]) -> T: ...   # constrained to int or float
def sort_key[T: Comparable](xs: list[T]) -> list[T]: ...  # T bound by a Protocol
```

## Protocols — structural typing

Prefer `Protocol` over ABCs/inheritance: any object with the right shape satisfies it, no explicit
subclassing. This keeps modules decoupled.

```python
from typing import Protocol

class Repo(Protocol):
    def get(self, id: int) -> User | None: ...
    def save(self, user: User) -> None: ...

def register(repo: Repo, user: User) -> None:   # accepts any structurally-matching object
    repo.save(user)
```

Use `@runtime_checkable` only when you must `isinstance()`-check (it checks method presence, not
signatures). Use ABCs instead of Protocols when you want to *share implementation*, not just a contract.

## Dataclasses

Default choice for data holders. Reach for `attrs` for validators/converters, Pydantic for
parse-and-validate at trust boundaries (API input, config).

```python
from dataclasses import dataclass, field

@dataclass(frozen=True, slots=True, kw_only=True)
class Config:
    host: str
    port: int = 8080
    tags: list[str] = field(default_factory=list)   # never a mutable default literal
```

- `frozen=True` → immutable & hashable. `slots=True` → less memory, no accidental attributes.
- `kw_only=True` → callers must name fields (safer against argument reordering).
- Use `field(default_factory=...)` for mutable defaults — never `tags: list = []`.

## Precise annotations toolkit

```python
from typing import Literal, Final, Annotated, TypedDict, NamedTuple

Mode = Literal["r", "w", "a"]           # closed set of values
MAX_RETRIES: Final = 3                   # cannot be reassigned (checked statically)

class RowTD(TypedDict):                   # dict with a fixed, typed shape
    id: int
    name: str
    email: str | None

class Point(NamedTuple):                  # immutable, typed, tuple-compatible
    x: float
    y: float

Port = Annotated[int, "1-65535"]         # attach metadata without changing the type
```

## Narrowing & exhaustiveness

Pair `match` with `assert_never` so adding a variant becomes a type error until handled.

```python
from typing import assert_never

def area(shape: Circle | Square) -> float:
    match shape:
        case Circle(r):
            return 3.14159 * r * r
        case Square(s):
            return s * s
        case _ as unreachable:
            assert_never(unreachable)   # type error here if a new shape is added
```

Use `TypeIs` (3.13; `TypeGuard` on older) for custom narrowing functions.

## Rules

- No `Any` without a comment justifying it; prefer `object` + narrowing, or a precise union.
- Type public boundaries strictly; let inference handle obvious locals.
- `Protocol` for contracts, ABC for shared implementation, dataclass for data.
- Run `pyright`/`mypy --strict` in CI — untyped code silently rots.
