# Python Async & Concurrency

## Pick the model by workload

| Workload | Use | Why |
|----------|-----|-----|
| Many concurrent I/O ops (network, disk, DB) | `asyncio` | One thread, cooperative; scales to thousands of sockets |
| Blocking I/O in a sync library | threads (`ThreadPoolExecutor`, `asyncio.to_thread`) | Releases the GIL during I/O |
| CPU-bound (parsing, math, compression) | `multiprocessing` / `ProcessPoolExecutor` | Sidesteps the GIL across processes |

**The GIL:** on standard CPython, only one thread runs Python bytecode at a time, so threads do *not*
speed up CPU-bound code — use processes. Python 3.13 ships an experimental free-threaded build
(`--disable-gil`); don't assume it in production yet.

## asyncio essentials

```python
import asyncio

async def fetch(client, url: str) -> bytes:
    async with client.get(url) as resp:      # async context managers for resources
        return await resp.read()

asyncio.run(main())                          # single entry point; creates & closes the loop
```

Never call a blocking function directly in a coroutine — it stalls the whole loop. Offload it:

```python
result = await asyncio.to_thread(blocking_call, arg)   # runs in the default thread pool
```

## Structured concurrency — prefer TaskGroup (3.11+)

`TaskGroup` bounds task lifetime to a scope and cancels siblings when one fails. Prefer it over bare
`create_task` (which leaks tasks) and over `gather` when you want fail-fast + guaranteed cleanup.

```python
async def load_all(urls: list[str]) -> list[bytes]:
    async with asyncio.TaskGroup() as tg:
        tasks = [tg.create_task(fetch(client, u)) for u in urls]
    return [t.result() for t in tasks]       # reached only if all succeeded
```

On failure the block raises an `ExceptionGroup`; handle with `except*`:

```python
try:
    await load_all(urls)
except* TimeoutError as eg:
    log.warning("timeouts: %d", len(eg.exceptions))
except* ValueError as eg:
    ...
```

`gather` is still fine for "run these and collect results/exceptions" without fail-fast:

```python
results = await asyncio.gather(*coros, return_exceptions=True)
```

## Timeouts and cancellation

```python
async with asyncio.timeout(5.0):     # 3.11+; cancels the block on expiry
    await slow_op()
```

- Cancellation is delivered as `CancelledError` raised at the next `await`. **Always re-raise it** —
  catching and swallowing it breaks cancellation semantics.

```python
try:
    await work()
except asyncio.CancelledError:
    await cleanup()
    raise                             # never swallow
```

## Bounding concurrency

Unbounded fan-out exhausts sockets/memory. Cap it with a semaphore:

```python
sem = asyncio.Semaphore(10)

async def guarded(url: str) -> bytes:
    async with sem:
        return await fetch(client, url)
```

Or use `asyncio.Queue` for producer/consumer pipelines.

## Threads & processes (when not asyncio)

```python
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor

with ThreadPoolExecutor(max_workers=8) as pool:      # I/O-bound blocking libs
    results = list(pool.map(download, urls))

with ProcessPoolExecutor() as pool:                  # CPU-bound
    results = list(pool.map(crunch, chunks))
```

For shared mutable state across threads use `threading.Lock`; prefer passing immutable data or using
queues over shared state.

## Rules

- One `asyncio.run()` entry point; don't create/close loops manually.
- Never block the event loop — offload sync work with `to_thread` / an executor.
- Prefer `TaskGroup` + `asyncio.timeout`; avoid orphan `create_task`.
- Always re-raise `CancelledError`.
- Bound concurrency (semaphore/queue) — never fan out unboundedly over external resources.
- Threads for I/O, processes for CPU — never threads for CPU-bound work on CPython.
