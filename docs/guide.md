# Nitrolite Documentation Guide

This document defines **how documentation should be written and structured** inside the Nitrolite repository.

Its purpose is to ensure that documentation:

* is consistent across the repository
* is easy for developers to navigate
* is easily retrievable by AI systems
* avoids duplication
* focuses on the most important information first

This guide must be followed when writing any documentation for the Nitrolite repository.

---

# 1. Documentation Principles

All documentation in the Nitrolite repository must follow these principles.

## 1.1 Single Source of Truth

Every piece of information must exist in **one canonical location**.

Other places may reference it but must not duplicate the content.

Example:

| Information           | Canonical location             |
| --------------------- | ------------------------------ |
| Protocol definitions  | `docs/protocol/terminology.md` |
| System architecture   | `docs/architecture/`           |
| Build Apps on Yellow Network          | `docs/build/`                  |
| Operator instructions | `docs/operator/`               |
| Code behaviour        | Go code comments               |

The website documentation repository must **reuse content from the main repository**, not redefine it.

---

## 1.2 AI-Friendly Documentation

Documentation should be written in a way that allows AI systems to reliably retrieve answers.

This requires:

* clear headings
* explicit terminology
* short conceptual sections
* clear definitions
* structured documents

Avoid narrative writing or long unstructured explanations.

---

## 1.3 Clear Separation of Concerns

Documentation must be separated into four main domains:

1. Protocol
2. System Architecture
3. Build (as in "build your applications on Yellow Network")
4. Operator

Each domain serves a different audience and must not mix responsibilities.

---

# 2. Documentation Structure

All documentation inside the repository must follow this directory structure.

```
docs/
    protocol/
    architecture/
    build/
    operator/
```

Each directory contains documentation for a specific domain.

---

# 3. Terminology Documentation

Terminology must be defined in a single canonical document.

Location:

```
docs/protocol/terminology.md
```

This document defines all protocol-level concepts.

Examples of concepts that belong here include:

* Channel
* State
* Epoch
* Settlement
* Operator
* Client

Each term must be defined once and used consistently across all documentation.

---

## Terminology Format

Each term must follow the same structure.

Example:

```
## Channel

Definition  
A channel is a state container shared between participants that allows
off-chain updates while maintaining on-chain security guarantees.

Purpose  
Channels enable fast off-chain execution while preserving the ability
to settle on-chain if necessary.

Used In  
- Channel lifecycle
- State updates
- Settlement
```

Terminology definitions must not contain implementation details.

---

# 4. Protocol Documentation

Protocol documentation describes **the system as a protocol**, independent of any specific implementation.

A reader must be able to implement the protocol from this documentation without reading the Nitrolite code.

Location:

```
docs/protocol/
```

Recommended documents:

```
overview.md
terminology.md
state-advancement.md
state-enforcement.md
```

---

## Protocol Documentation Must Include

Protocol documents must describe:

* protocol concepts
* state structures
* rules governing state transitions
* lifecycle of channels
* settlement and dispute behaviour
* interaction with blockchains

Protocol documentation must avoid:

* code references
* repository structure
* implementation details

## Language for Structures and Functions

Protocol documentation must use **language-neutral pseudocode** when describing structures or functions.

Use simple struct-like notation for data structures.

Use plain function signatures with named parameters and return types.

Rules:

* Do not use syntax specific to any programming language (Go, TypeScript, Solidity, etc.)
* Use CamelCase for field and function names
* Keep pseudocode minimal — only show what is needed to convey the concept

---

# 5. System Architecture Documentation

System architecture documentation explains **how the Nitrolite implementation realizes the protocol**.

Location:

```
docs/architecture/
```

Recommended documents:

```
system-overview.md
node-architecture.md
storage.md
networking.md
security.md
```

---

## Architecture Documentation Must Include

Architecture documentation must describe:

* system components
* internal services
* communication patterns
* storage model
* security mechanisms
* how the protocol is implemented

Architecture documentation may reference code modules.

Architecture documentation must not redefine protocol rules.

---

# 6. Separating Protocol and Architecture

The protocol and architecture documentation may appear similar because the protocol was developed together with the implementation.

However they must remain conceptually separate.

### Protocol answers

```
What are the rules of the system?
```

### Architecture answers

```
How does Nitrolite implement those rules?
```

Example:

| Topic             | Protocol                          | Architecture                                 |
| ----------------- | --------------------------------- | -------------------------------------------- |
| State             | Defines state structure and rules | Explains how state is stored                 |
| Settlement        | Defines settlement process        | Explains which component executes settlement |

Protocol documentation must remain **implementation-independent**.

Architecture documentation describes **Clearnet specifically**.

---

# 7. "Build Apps on Yellow Network" Documentation

Build documentation must onboard developers to start building on top of Yellow Network with minimum friction. This documentation must highlight only protocol concepts and SDK methods necessary for app developers. It must not describe protocol internals.


Location:

```
docs/build/
```

Recommended documents:

```
overview.md
app.md
develop.md
examples.md
```



1. **app.md** must cover:

* how to register an app
* app session lifecycle
* concept of daily allowances
* app session keys

2. **develop.md** must list SDK methods necessary for app development.
3. **examples.md** must show real-world use case examples of application flows built with the SDK. Starting with simplest examples and gradually increasing complexity.


---

# 8. Operator Documentation

Operator documentation explains how to run and maintain infrastructure.

Location:

```
docs/operator/
```

Recommended documents:

```
running-node.md
configuration.md
monitoring.md
upgrades.md
```

---

## Operator Documentation Must Include

Operator documentation must cover:

* node deployment
* configuration parameters
* operational procedures
* monitoring requirements
* upgrade procedures

Operator documentation must not include protocol explanations.

---

# 9. Document Structure Requirements

All documents must follow a predictable structure.

This ensures both developers and AI systems can quickly locate information.

---

## README.md Structure

Every repository README must contain the following Header, followed with flexible component-specific sections.

```
# Project Name

Short description of the project.

## Overview

High level explanation of the system.

## Documentation

Links to detailed documentation.
```

---

## Overview Document Structure

Overview documents must contain:

```
# Overview

## Purpose

Why this component exists.

## Concepts

Key ideas required to understand it.

## How It Works

Explanation of behaviour and interactions.

## Table of contents

Bulletpoints with links to documentation and short descriptions.
```

---

# 10. Writing Requirements

All documentation must follow these writing rules.

### Use precise terminology

Always use defined protocol terms.

### Avoid ambiguity

Explain behaviour explicitly.

### Avoid implementation leakage in protocol docs

Protocol documentation must not reference code.
