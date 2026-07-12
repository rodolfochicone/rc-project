# Python Testing (pytest)

`pytest` is the default. Plain functions, `assert` with introspection — no `unittest` class boilerplate.
Tests must encode *why* the behavior matters (see rc-tdd), not just pin current output.

## Layout & discovery

```
tests/
  conftest.py          # shared fixtures, auto-discovered by pytest
  test_orders.py       # files: test_*.py; functions: test_*
```

```python
def test_total_excludes_cancelled_items():
    order = Order(items=[Item(10), Item(5, cancelled=True)])
    assert order.total() == 10      # WHY: cancelled items must not bill the customer
```

## Parametrize — the table-driven pattern

One test, many cases; each case reports separately.

```python
import pytest

@pytest.mark.parametrize(
    ("raw", "expected"),
    [
        ("1.5", 1.5),
        ("0", 0.0),
        ("-2", -2.0),
    ],
)
def test_parse_amount(raw: str, expected: float):
    assert parse_amount(raw) == expected

@pytest.mark.parametrize("bad", ["", "abc", "1,5"])
def test_parse_amount_rejects_garbage(bad: str):
    with pytest.raises(ValueError):
        parse_amount(bad)
```

## Fixtures

Setup/teardown as dependencies. `yield` splits setup from cleanup; `scope` controls lifetime
(`function` default, `module`, `session`).

```python
@pytest.fixture
def db():
    conn = connect(":memory:")
    conn.migrate()
    yield conn                    # test runs here
    conn.close()                  # teardown, even on failure

def test_insert(db):
    db.insert(User("a"))
    assert db.count() == 1
```

Prefer fixtures over module-level globals; put shared ones in `conftest.py`.

## Mocking & patching

Patch where the name is *used*, not where it's defined.

```python
def test_notify_calls_gateway(mocker):        # pytest-mock's `mocker`
    send = mocker.patch("app.orders.send_sms")   # app.orders imported send_sms
    place_order(order)
    send.assert_called_once_with(order.phone, mocker.ANY)
```

`monkeypatch` for env/attrs:

```python
def test_reads_region(monkeypatch):
    monkeypatch.setenv("AWS_REGION", "sa-east-1")
    assert config().region == "sa-east-1"
```

## Async tests

Use `anyio` or `pytest-asyncio`:

```python
@pytest.mark.anyio
async def test_fetch_returns_body():
    body = await fetch(client, "https://example.test")
    assert body
```

## Exceptions & approximate values

```python
with pytest.raises(ValueError, match="port"):    # asserts message too
    Config(port=-1)

assert result == pytest.approx(0.1 + 0.2)         # float comparison
```

## Property-based (hypothesis)

For invariants over a large input space:

```python
from hypothesis import given, strategies as st

@given(st.lists(st.integers()))
def test_sort_is_idempotent(xs: list[int]):
    assert sorted(sorted(xs)) == sorted(xs)
```

## Coverage

```
pytest --cov=app --cov-report=term-missing
```

Target ≥80%, but coverage measures execution, not assertion quality — a line run without a meaningful
assert is still untested. Cover branches and error paths, not just the happy path.

## Rules

- `test_*` functions + `assert`; no `unittest.TestCase` unless a codebase already uses it.
- `parametrize` the case table; don't copy-paste near-identical tests.
- Fixtures with `yield` for setup/teardown; shared ones in `conftest.py`.
- Patch at the point of use; assert on calls/messages, not just "no exception".
- Test error paths and boundaries; a test that can't fail when logic changes is worthless (Rule 7).
