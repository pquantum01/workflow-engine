# Requirements - Workflow Engine

## Overview

A standalone workflow engine (like Jira) that allows users to define processes, workflows, tasks, and state machines with full configurability. Processes run independently with unique IDs and can span days, driven by humans, AI agents, rules, or other processes.

## Core Concepts

### Process Definition
- A **template** that describes a type of work (e.g., "Bug Triage", "Employee Onboarding").
- Contains one or more **Workflow Definitions**.
- Admin-configurable via UI.

### Process Instance
- A **running instance** of a Process Definition.
- Has a unique `process_id`.
- Can run for days/weeks.
- Many instances can run in parallel.
- Supports **parent-child** hierarchy (exactly 1 level deep: parent -> children, no grandchildren).

### Workflow Definition
- Belongs to a Process Definition.
- Has a **State Machine** defining valid states and transitions.
- Contains **Task Definitions** that drive state transitions.

### Workflow Instance
- A running instance of a Workflow Definition, tied to a Process Instance.
- Tracks current state within its state machine.

### Task Definition
- Belongs to a Workflow Definition.
- Configurable completion rules:
  - **Single task completion** triggers a workflow state transition.
  - **Multiple tasks** can contribute to a single state transition (all must complete, or N-of-M, etc.).

### Task Instance
- A running instance of a Task Definition, tied to a Workflow Instance.
- Has its own status (pending, in_progress, completed, skipped, failed).
- Can be completed by: human action, AI agent, rule evaluation, or another process.

### State Machine
- Fully configurable per Workflow Definition.
- Defines: states, transitions, initial state, terminal states.
- Transitions can be guarded by task completion rules.

## Functional Requirements

### FR-1: Process Definition Management
- FR-1.1: CRUD operations for Process Definitions.
- FR-1.2: Each Process Definition has a name, description, version, and status (draft/active/archived).
- FR-1.3: A Process Definition contains one or more Workflow Definitions.

### FR-2: Workflow Definition Management
- FR-2.1: CRUD operations for Workflow Definitions within a Process Definition.
- FR-2.2: Each Workflow Definition has a name, description, and a State Machine configuration.
- FR-2.3: A Workflow Definition contains zero or more Task Definitions.

### FR-3: State Machine Configuration
- FR-3.1: Define states with name and metadata.
- FR-3.2: Define transitions between states.
- FR-3.3: Designate initial state and terminal states.
- FR-3.4: Transitions can have **completion rules** — conditions based on task completions that must be satisfied.

### FR-4: Task Definition Management
- FR-4.1: CRUD operations for Task Definitions within a Workflow Definition.
- FR-4.2: Each Task Definition has a name, description, and type.
- FR-4.3: Configure how task completions map to state transitions:
  - Single task -> triggers transition.
  - Multiple tasks -> all must complete to trigger transition.
  - N-of-M tasks -> at least N of M specified tasks must complete.

### FR-5: Process Instance Lifecycle
- FR-5.1: Start a new Process Instance from a Process Definition.
- FR-5.2: Each instance gets a unique `process_id`.
- FR-5.3: Run many instances in parallel.
- FR-5.4: Process instances can span days.
- FR-5.5: State changes can be triggered by: human, AI agent, rule, or another process.

### FR-6: Parent-Child Process Hierarchy
- FR-6.1: A Process Instance can have child Process Instances.
- FR-6.2: Maximum 1 level of hierarchy (parent -> children only).
- FR-6.3: Parent can query status of children.
- FR-6.4: Child completion can optionally trigger parent state transitions.

### FR-7: Admin Portal (UI)
- FR-7.1: Configure Process Definitions, Workflow Definitions, Task Definitions, State Machines.
- FR-7.2: Visual state machine editor.
- FR-7.3: Kanban-like view to browse running Process Instances.
- FR-7.4: Filter/search/sort Process Instances.
- FR-7.5: Drill into a Process Instance to see Workflow and Task status.

## Non-Functional Requirements

### NFR-1: Performance
- Support 1000+ concurrent Process Instances.
- API response time < 200ms for standard operations.

### NFR-2: Reliability
- State transitions must be atomic and consistent.
- No data loss on crashes (persistent storage).

### NFR-3: Testability
- All backend packages must have unit tests.
- Integration tests for API endpoints and state machine transitions.
- Frontend components must have unit tests.

### NFR-4: Extensibility
- Plugin-friendly architecture for custom task executors.
- Webhook support for external integrations.

## API Actors

| Actor | Description |
|-------|-------------|
| Admin | Configures Process/Workflow/Task Definitions and State Machines |
| Operator | Starts Process Instances, monitors progress, acts on tasks |
| System | AI agents, rules, or other processes that trigger state changes |
