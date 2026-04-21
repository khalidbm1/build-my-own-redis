# Build My Own Redis (Please Don't Use This In Production)

> _"Why pay for a reliable, battle-tested, millisecond-precise in-memory data store
> written by geniuses over a decade when you could build a wobbly imitation yourself
> in a weekend and learn a humbling amount about `bufio.Reader` in the process?"_
>
> — probably me, at 2 AM

Welcome to **Build My Own Redis**, a Go project whose main achievement is proving
that `redis-cli` will happily talk to anything that speaks RESP with enough
conviction.

---

## What Is This

An in-memory key-value server that:

- Speaks RESP (Redis Serialization Protocol) well enough to fool `redis-cli`.
- Stores strings, lists, and sets.
- Expires keys (eventually).
- Writes to an AOF file so your data survives restarts — assuming the AOF file
  itself survives, which is not a given.
- Benchmarks at ~**101,000 req/s** on an M-series MacBook Air, which sounds
  impressive until you realize real Redis does ~1,000,000+ req/s and also
  doesn't log every single connection to stdout like a caffeinated newsreader.

## Why

Because the official Redis source code is 200k+ lines of C and I am one person
with a terminal and emotional baggage.

## Architecture

```text
┌───────────────┐    RESP    ┌──────────────┐
│   redis-cli   │ ─────────► │  our server  │
└───────────────┘            └──────┬───────┘
                                    │
                         ┌──────────┴─────────┐
                         │                    │
                   ┌─────▼──────┐      ┌──────▼──────┐
                   │  Store     │      │    AOF      │
                   │ (map + mu) │      │  (file++)   │
                   └─────┬──────┘      └─────────────┘
                         │
                   ┌─────▼─────────────┐
                   │ expiry goroutine  │
                   │ (sweeps every 10s,│
                   │  judges your TTLs)│
                   └───────────────────┘
```

Yes, it is just a `map[string]*Entry` behind a `sync.RWMutex`. No, I don't want
to hear about lock-free data structures right now.

## Supported Commands

Tier S (the ones that actually work):

- `PING`, `ECHO`
- `SET` (with `EX`/`PX`!), `GET`, `DEL`, `EXISTS`, `KEYS *`
- `INCR`, `DECR`
- `LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LLEN`, `LRANGE`
- `SADD`, `SREM`, `SMEMBERS`, `SISMEMBER`, `SCARD`
- `EXPIRE`, `PEXPIRE`, `TTL`, `PERSIST`

Tier F (the ones I definitely meant to add later):

- Pub/Sub
- Transactions
- Replication
- Clustering
- Anything that would make your DBA sleep at night

## Quick Start

```bash
make build
./bin/redis-server
```

In another terminal:

```bash
redis-cli PING
# PONG

redis-cli SET vibes immaculate EX 60
# OK

redis-cli GET vibes
# immaculate
```

Yes. `redis-cli`. The real one. It does not know. Please do not tell it.

## Running The Tests

```bash
./test.sh
```

Twenty tests. All green. Cope.

## Running The Benchmark

```bash
make benchmark
```

Expected output when I first wrote the Makefile (aspirational):

```text
SET: 50000.00 requests per second
GET: 50000.00 requests per second
```

Actual output (not aspirational):

```text
GET: 101419.88 requests per second, p50=0.271 msec
```

I'd like to take credit for this, but honestly it's mostly because Go's net
package is absurdly good and my Mac has too much RAM.

## Docker

```bash
make docker-build
make docker-run
```

It works. I think. I watched it start up and I didn't check.

## Known "Features"

- **Logs everything.** Every command, every connection, every disconnect.
  At 100k ops/sec this produces approximately one small rainforest of stdout
  per benchmark run. There is a TODO about gating this behind a debug flag.
  The TODO has been there a while. The TODO has made friends with other TODOs.

- **Expiry is lazy-ish.** Expired keys are cleaned up every 10 seconds by a
  background goroutine, or opportunistically on `GET`. If you `TTL` a key the
  exact microsecond it expires you might observe a quantum superposition.
  That's a feature. In physics.

- **The AOF replay is `go commands.Dispatch(...)`.** Meaning: restart replays
  writes through the exact same code path as live traffic. Elegant? Yes.
  Slightly haunted? Also yes.

- **No authentication.** If you `EXPOSE 6379` on the internet, you are going
  to have a bad time, and frankly you deserve it.

## File Tour

```text
cmd/redis-server/       # main(). The bit that cries SIGTERM.
internal/protocol/      # RESP reader/writer. `bufio` all the way down.
internal/storage/       # The map. That's it. It's just a map.
internal/commands/      # Dispatch table. If-else but organized.
internal/persistence/   # AOF. "Save? Who said anything about saving?"
internal/server/        # TCP listener + connection lifecycle.
```

## Things I Learned

1. `bufio.Reader.ReadByte` is your best friend until it isn't.
2. `interface{}` is what happens when you want generics but don't want to
   think about them yet.
3. Graceful shutdown is 40% of the code and 100% of the regret.
4. If you return early before `WaitGroup.Wait()`, the WaitGroup waits for you
   in your dreams instead.
5. `os.O_APPEND` without `os.O_WRONLY` is a Go koan: a file descriptor that
   exists, opens, and refuses to be written to. Meditate on this.

## Contributing

Don't. Actually please do. But know that this repository is a learning exercise
wearing a trench coat pretending to be a database.

## License

You can use this for anything. I legally cannot stop you. I would strongly
prefer you use actual Redis.

## Credits

- Salvatore Sanfilippo, for making Redis.
- Whoever designed RESP, for making it parseable by a person who has not
  slept.
- `redis-cli`, an unwitting accomplice.
- Me, for typing.
