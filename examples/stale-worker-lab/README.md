# The lease expired. The goroutine didn't.

A small Go experiment about stale workers, optimistic concurrency, and ownership
fencing. Independent educational example; it does not describe an employer or
client incident. No credentials, cluster, database, or third-party Go packages.

## The failure worth reproducing

Worker A reads a configuration and decides what to write. A pauses. Its lease
expires, B becomes owner, and B applies a newer decision. A resumes.

A version check can reject A's original write. But if A handles the conflict by
fetching the latest version **while keeping its old decision**, its next write
can succeed. A fresh version is not evidence that the intent is still valid.

```mermaid
sequenceDiagram
    participant A as Worker A
    participant T as Write target
    participant B as Worker B
    A->>T: Read version 0; compute old intent
    Note over A: Paused; ownership changes
    B->>T: Install epoch 2; write new intent
    A->>T: Old intent with version 0
    T-->>A: Version conflict (unfenced mode)
    A->>T: Fetch version 1; replay old intent
    T-->>A: Accepted without fencing; rejected with fencing
```

## Run it

From this directory, with Go 1.26:

```sh
go run .
go test -race ./...
go vet ./...
```

The unfenced run ends with `A: obsolete desired configuration`. The fenced run
ends with `B: current desired configuration`, rejecting both attempts by A.
The takeover is an explicit deterministic interleaving, not a sleep-based race.
The test suite also exercises concurrent writes within the same ownership epoch.

## The actual guarantee—and the expensive assumption

This model uses a **strict active epoch** installed at the protected write target
by a trusted ownership transition. The target validates epoch and version in the
same critical section as the mutation. That rejects A even before B's first data
write. Writers cannot choose a larger token and declare themselves owner.

This is stronger than a target that only remembers the highest token seen on
data writes: that pattern may still accept A until B's higher token arrives.
If ownership and data live in separate systems, atomically installing and
enforcing the fence is a design problem, not something this mutex solves.

| Mechanism | What it addresses | What it does not establish |
| --- | --- | --- |
| Lease | Temporary coordination | That a paused process stopped running |
| Version check / CAS | Changes since a particular read | Whether a refreshed old decision is still valid |
| Active epoch at target | Authority of the writer at mutation time | That its business decision is correct |
| Context cancellation | Cooperative stopping | Revocation of an already dispatched side effect |
| Idempotency key | Duplicate execution of the same operation | Authority of a stale owner |

## Kubernetes connection

Kubernetes `resourceVersion` is an opaque concurrency token; it is **not** the
integer ownership epoch used in this example. A conflict retry must re-read and
reconcile current desired state. Acquiring a Kubernetes Lease does not install
this example's fence on arbitrary ResourceQuota writes or external services.
Reading a Lease immediately before a write still leaves a check/write race.

An actual design needs to specify the ownership authority, fence installation,
atomic validation at each side-effect destination, and behavior during failures.
If a destination cannot enforce a fence, do not claim fencing protection there.

## Limits and next experiments

This is an in-memory model, not a distributed lock library or Kubernetes
controller. It does not model crashes, durable epoch allocation, lease clocks,
network partitions, token authentication, or multi-resource transactions. There
are no performance claims. Useful follow-ups are a transactional database
implementation and a real controller conflict test against an API server.

References: [Kubernetes API concurrency](https://kubernetes.io/docs/reference/using-api/api-concepts/)
and [distributed locks and fencing](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html).
