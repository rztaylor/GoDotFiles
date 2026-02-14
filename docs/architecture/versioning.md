# Schema Versioning Strategy

## Context
As GDF evolves, the structure of its configuration files (apps, profiles, global config) will likely change. To ensure stability and enable automated upgrades, we need a robust versioning strategy for these YAML files.

## Decision
We will use a **combined `kind` field** to denote both the resource type and its schema version.

### Format
The `kind` field must follow the format: `<Type>/<Version>`

-   **Type**: The resource type (e.g., `App`, `Profile`, `Config`, `Recipe`).
-   **Version**: The schema version, strictly following `v` + number (e.g., `v1`, `v2`).

**Example:**
```yaml
kind: App/v1
name: my-app
description: A sample application
```

### Why this approach?
1.  **Simplicity**: Combining type and version into a single field reduces boilerplate in user-authored files. It avoids the verbosity of Kubernetes-style `apiVersion` + `kind` + `metadata` structures.
2.  **Explicit Versioning**: It forces every file to declare its version, preventing ambiguity.
3.  **No Legacy Support**: Since GDF has not yet had a public release, we are choosing *not* to support unversioned "legacy" files. This simplifies the codebase by removing the need for heuristic detection of old formats.

## Implementation Details
1.  **Parsing**: The system will peek at the `kind` field before full unmarshalling.
2.  **Validation**: Files missing the `kind` field or using an invalid format will be rejected with a clear error message.
3.  **Upgrades**: When a new version is introduced (e.g., `v2`), an "Upgrader" component can be registered to transform `v1` data into the `v2` struct structure in memory.

## Future Proofing
If we need to introduce breaking changes to the `App` structure, we will:
1.  Define a new `AppV2` struct (or update the existing on if backwards compatible in Go).
2.  Increment the version to `App/v2`.
3.  Implement a translation layer to convert `App/v1` YAML input to the internal `App` representation.
