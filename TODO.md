# ERC-8004 Integration — Task Tracker

> Update this file after completing each task. Change status from `⬜ TODO` → `🔄 IN PROGRESS` → `✅ DONE` → `❌ BLOCKED`.
> Add notes with PR links, issues encountered, or decisions made.

## Shared Contracts (Sepolia, Chain ID: 11155111)

| Contract | Address |
|----------|---------|
| AgentRegistry | `0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3` |
| HackathonVault | `0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90` |
| RiskRouter | `0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC` |
| ReputationRegistry | `0x423a9904e39537a9997fbaF0f220d79D7d545763` |
| ValidationRegistry | `0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1` |

---

## Task Status

### Phase 1: Foundation

| # | Task | Status | Assignee | Notes |
|---|------|--------|----------|-------|
| 1 | Add go-ethereum dependency and update configuration | ✅ DONE | opencode | |
| 2 | Extract contract ABIs and generate Go bindings | ✅ DONE | opencode | Depends on: 1 |

### Phase 2: Blockchain Client Core

| # | Task | Status | Assignee | Notes |
|---|------|--------|----------|-------|
| 3 | Create blockchain client core (`client.go` + `config.go`) | ✅ DONE | opencode | Depends on: 2 |
| 4 | Implement agent registration and vault operations | ⬜ TODO | | Depends on: 3 |
| 5 | Implement EIP-712 signing utilities | ⬜ TODO | | Depends on: 3 |
| 6 | Implement trade submission to RiskRouter | ⬜ TODO | | Depends on: 3, 5 |
| 7 | Implement validation checkpoint posting | ⬜ TODO | | Depends on: 3 |
| 8 | Implement reputation registry interactions | ⬜ TODO | | Depends on: 3 |

### Phase 3: Trade Executor

| # | Task | Status | Assignee | Notes |
|---|------|--------|----------|-------|
| 9 | Create trade executor (`internal/executor/`) | ⬜ TODO | | Depends on: 4, 5, 6, 7, 8 |

### Phase 4: Integration

| # | Task | Status | Assignee | Notes |
|---|------|--------|----------|-------|
| 10 | Add blockchain state persistence (SQLite + InfluxDB) | ⬜ TODO | | Depends on: 9 |
| 11 | Wire blockchain client into application startup | ⬜ TODO | | Depends on: 9, 10 |
| 12 | Add blockchain API endpoints | ⬜ TODO | | Depends on: 11 |
| 13 | Add CLI commands for agent lifecycle | ⬜ TODO | | Depends on: 11 |

### Phase 5: Hardening

| # | Task | Status | Assignee | Notes |
|---|------|--------|----------|-------|
| 14 | Add error handling, retry logic, and circuit breaker | ⬜ TODO | | Depends on: 12, 13 |
| 15 | Write integration tests | ⬜ TODO | | Depends on: 14 |

---

## Dependency Graph

```
Task 1 (deps/config)
  └── Task 2 (ABI bindings)
        └── Task 3 (client core)
              ├── Task 4 (registration/vault)
              ├── Task 5 (EIP-712 signing)
              │     └── Task 6 (trade submission)
              ├── Task 7 (validation checkpoints)
              └── Task 8 (reputation)
                    └── Task 9 (trade executor)
                          ├── Task 10 (persistence)
                          └── Task 11 (app wiring)
                                ├── Task 12 (API endpoints)
                                └── Task 13 (CLI commands)
                                      └── Task 14 (error handling)
                                            └── Task 15 (integration tests)
```

**Parallelizable after Task 3:** Tasks 4, 5, 7, 8
**Parallelizable after Task 11:** Tasks 12, 13

---

## Progress Summary

| Phase | Tasks | Done | Remaining |
|-------|-------|------|-----------|
| 1. Foundation | 1-2 | 2 | 0 |
| 2. Blockchain Client Core | 3-8 | 1 | 5 |
| 3. Trade Executor | 9 | 0 | 1 |
| 4. Integration | 10-13 | 0 | 4 |
| 5. Hardening | 14-15 | 0 | 2 |
| **Total** | **1-15** | **3** | **12** |

---

## Change Log

| Date | Task | Agent | Change |
|------|------|-------|--------|
| 2026-04-05 | 1 | opencode | Added go-ethereum dependency, updated .env.example with blockchain config, added config defaults |
| 2026-04-05 | 2 | opencode | Created ABI JSON files and generated Go bindings for all 5 contracts |
| 2026-04-05 | 3 | opencode | Created blockchain client core with config, RPC connection, wallet parsing, contract bindings |

---

## Notes & Decisions

<!-- Add architectural decisions, blockers, or important context here -->

---

## How to Update

When an agent completes a task:

1. Change the status in the table: `⬜ TODO` → `✅ DONE`
2. Add the agent name in the Assignee column
3. Add a row to the Change Log with date, task number, agent name, and summary
4. Update the Progress Summary counts
5. If blocked, change to `❌ BLOCKED` and explain in Notes & Decisions
