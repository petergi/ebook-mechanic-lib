# EBMLib Architecture (Modular Monolith Baseline)

*Phase:* Architecture Baseline Locked  
*Last Updated:*

This document reflects the **current, implemented** architecture of EBMLib. It captures what exists in the codebase today (Go modular monolith, and provides diagrams plus a quick reference. Cross-references to ADRs are included for traceability.

## 1. Current State Snapshot

- **Runtime:** 

- **Domain services:**

- **Data layer:** 

- **API Layer:**

- **Observability:**

- **Deployment:**

- **Testing:** 

  

## 2. Architecture Overview

```mermaid
flowchart LR

```

### Notable Behaviors

- 



## 3. Data Model (ERD)

```mermaid
erDiagram

```



## 4. API Surface

```yaml

```

## 5. Project & Task Systems (Implementation Overview)
- **Services:** `internal/services/project_service.go`, `internal/services/task_service.go`. Responsibilities: project/task CRUD, project-scoped task listings, pagination, JSONB metadata, status/priority/due date handling, audit logging.
- **APIs:**
- **Data model:** 
- **Status:** 
- 



## 9. Roadmap & Deferred Items

- 

