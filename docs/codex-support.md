---
status: living-plan
linear: CF-342
scope: Add Codex support without disrupting Claude Code users
intent: Track checkpoints, invariants, risks, and decisions for the multi-phase Codex support work.
last_reviewed: 2026-05-12
---

# Codex Support Plan

This document tracks the incremental path to Codex support. It is intentionally broader than any single PR, but each checkpoint must remain small enough to verify without changing existing Claude Code behavior.

## Core Invariant

Phase 2 must not change any installed Claude Code hook command string, settings file location, environment variable, backend request body, daemon state filename, inbox JSON shape, or user-facing default behavior.

In particular, existing Claude Code users must continue to use commands such as:

- `confab hook session-start`
- `confab hook session-end`
- `confab hook pre-tool-use`
- `confab hook post-tool-use`
- `confab hook user-prompt-submit`

## Current Phase: Claude Provider Extraction

Goal: extract Claude-specific local behavior into a concrete provider package without implementing Codex.

Non-goals for this phase:

- No Codex provider stub.
- No `--tool` CLI flag.
- No backend `tool_name` payload.
- No daemon state or inbox schema change.
- No transcript normalization.
- No Codex hook config writer.
- No skill abstraction for `/til` or `/retro`.
- No generic normalized hook input model.

Checklist:

- [ ] Add `pkg/provider` with concrete `ClaudeCode`.
- [ ] Move Claude path/settings/session-root knowledge behind `ClaudeCode` methods.
- [ ] Move Claude hook input parsing behind concrete `ClaudeCode` methods.
- [ ] Move Claude parent process matching/detection behind concrete `ClaudeCode` methods.
- [ ] Rename hook request/response Go types to Claude-specific names while preserving JSON wire shape.
- [ ] Keep existing exported Claude-compatible wrappers where callers rely on them.
- [ ] Add fixture tests proving installed hook JSON remains unchanged.
- [ ] Add response tests proving Claude hook JSON output remains unchanged.
- [ ] Keep `CONFAB_CLAUDE_DIR` as the only state-dir override.
- [ ] Keep temporary compatibility shims (`pkg/config/paths.go`, `pkg/discovery/hook.go`) until provider call sites settle.

## Later Checkpoints

- [ ] Backend `tool_name` support: additive request field, backend default for legacy clients, dedup by `(user_id, tool_name, external_id)`.
- [ ] Cleanup compatibility shims after provider ownership is stable: move remaining path and hook parsing callers directly to provider APIs, then remove wrappers that no runtime code needs.
- [ ] CLI provider selection: introduce `--provider claude-code|codex` surgically on commands with real provider-specific behavior.
- [ ] Codex provider: implement real Codex paths, rollout discovery, hook payload parsing, and hook config writing from current Codex docs/source.
- [ ] Codex daemon behavior: run the real daemon lifecycle against Codex rollout files, but route backend calls to a local dry-run backend until backend support exists.
- [ ] Transcript normalization: add backend and frontend normalization keyed by tool name before enabling analytics/Smart Recap for Codex.
- [ ] Codex subagents: quick-follow TODO. Model separate rollout files and parent relationships from `thread_source=subagent`, `agent_path`, `agent_role`, and `agent_nickname`.
- [ ] Skills: revisit `/til` and `/retro` separately; Claude slash-command skills should remain Claude-specific until Codex has a well-defined surface.

## Decisions

- Provider work starts as concrete Claude extraction, not a premature multi-provider abstraction.
- Hook payload formats are provider-specific. Do not introduce a generic normalized hook input until Codex requirements are confirmed.
- `ClaudeSettings` remains Claude-specific because it wraps `~/.claude/settings.json`.
- Parent PID monitoring remains Claude-specific implementation detail for now.
- `/til` and `/retro` remain Claude-specific for this phase.
- Documentation visible to users should remain Claude-specific until Codex support is real.
- Codex support starts CLI-first but includes the full local lifecycle: discovery, `list`, `save`, daemon dry-run sync, and hook installation.
- Codex must not make real backend API calls in this phase. Dry-run calls log local operation metadata to the main Confab log and return mocked responses.
- Codex session identity is parsed from rollout filenames matching `rollout-<timestamp>-<uuid>.jsonl`.
- Codex rollout `session_meta` is parsed for metadata and top-level filtering. `confab list --provider codex` includes user sessions only: missing/`user` `thread_source`, and no `agent_path`, `agent_role`, or `agent_nickname`.
- Codex local discovery reads rollout JSONL files only. Do not read Codex SQLite state in the first Codex CLI slice.
- Codex hook install should match Claude's seamless setup posture: preserve existing user config, make backups, install idempotently, enable `features.codex_hooks = true`, and clearly surface that feature flag change in CLI output.
- Codex hooks should use existing handler shapes with explicit provider selection, e.g. `confab hook session-start --provider codex`.
- Provider selection flags should be added only where they have real behavior.
- Daemon state should be provider-aware going forward, while preserving legacy Claude state file lookup and cleanup for existing users.

## Compatibility Shims (Future Cleanup)

These exist only to keep this checkpoint's diff focused. They should be removed in a later checkpoint, once provider usage settles and Claude behavior has not regressed:

- `pkg/discovery/hook.go` — `ReadHookInputFrom` now forwards to `provider.ClaudeCode{}.ReadSessionHookInput`. Runtime callers have all moved to the provider directly; only `pkg/discovery/hook_test.go` still exercises this wrapper. Remove after one checkpoint of bake time; the `..`-traversal assertion is already covered in `pkg/provider/claude_test.go`.
- `pkg/config/paths.go` — `GetClaudeStateDir`, `GetProjectsDir`, `GetClaudeSettingsPath`, and the `ClaudeStateDirEnv` constant all forward to `provider.ClaudeCode{}`. Real callers (`cmd/skills.go`, `pkg/config/skill_til.go`, `pkg/config/skill_retro.go`, `pkg/discovery/sessions.go`) should call `provider` directly once the skill and discovery surfaces are moved.

## Risks

- Mechanical hook type renames can hide JSON wire changes. Protect with exact response and hook settings tests.
- Provider constructor injection can sprawl. Limit command constructor changes to touched hook/status flows.
- Daemon state and inbox files are operationally sensitive. Do not change their filenames or JSON shape in this phase.
- Codex assumptions can drift quickly. Confirm Codex hook config, transcript layout, and subagent metadata before implementing the Codex provider.
