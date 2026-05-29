# Event Store

A small HTTP service that stores events in an append-only log and reads them back by ID, even after a restart.

## Setup

### Install & Start
```bash
go run .
```

### API Commands
```bash
# POST an event
curl -X POST http://localhost:8080/events   -H "Content-Type: application/json"   -d '{"name": "daniel", "action": "login"}'

# GET an event
curl http://localhost:8080/events/bb7769cd-9f23-464c-a3e6-3584cf000a4f

# GET stats
curl http://localhost:8080/stats
```

---

## Architecture

![Read & Write Architecture](eventstore_architecture.svg)

### How a write works
A write works by going to the end of the file using `Seek`, since it's an append-only log, then the the content is written as a string
then forcefully saved to disk with `e.file.Sync`. The write uses a mutex lock to serialize writes to prevent race conditions 

### How a read works
A read works by first checking the index to see if the event exists, then goes on to read the exact bytes for that particular index that exists since we store both offset and legnth of content in bytes.

---

## Core Concepts

### Why append-only is safer than overwriting
Append-only is safer compared to overwriting because if the process crashes while overwriting the data is now corrupted, whereas in append-only a crash mid-write only leads to an incomplete entry.

### Why an index makes reads fast
Indexes make read fast because they provide a means to skip full scans and jump straight to the requested data.
Although they come at a tradeoff: every index can improve reads but slow down writes.

---

## Recovery
![alt text](image.png)

---

## What I Struggled With
Verifying that unicode byte counting was correct.
Understanding why I needed both offset and length

---

## What I Learned
This is the simplest version of a storage engine: an append-only log with hash index. It has some problems which are then addressed by advanced storage engines like LSM and B-Trees. But they all have something this in common as a crash recovery strategy.
This thinking also translates to distributed systems where crashes can happen, one key to recovery is to record the intent to do something before doing it. That's the idea behind the WAL, outbox pattern and so on.

---

## Resources
Designing Data-Intensive Applications - Martin Kleppmann; Chapter 3 - Database Storage Engines

---

## Why This Made Me a Better Backend Developer
This has made me develop a strategy for crash recovery and I've also seen how understanding how low-level concepts can have various use cases when applied to different context.
