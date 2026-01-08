
# Software Requirements Specification (SRS)
## Kubernetes Multi-Cluster Namespace Configuration Comparison Tool

---

## 1. Introduction

### 1.1 Purpose
This document specifies the requirements for a tool that compares Kubernetes namespace configurations across multiple clusters. One namespace (baseline) is selected as the standard, and other namespaces are compared against it to detect configuration drift.

### 1.2 Scope
The system provides:
- Multi-cluster namespace selection
- Resource-level configuration comparison
- Visual diff reporting via tabular UI
- YAML-based application configuration
- Extensible comparison modes (Mode A initially)

### 1.3 Definitions
- **Baseline**: The reference namespace used as comparison standard
- **Target**: A namespace compared against the baseline
- **Mode A**: Structural configuration diff
- **Mode B**: Policy/invariant validation (future)

---

## 2. Overall Description

### 2.1 Product Perspective
The tool operates as a standalone web application or CLI-backed service interacting with Kubernetes APIs using kubeconfig contexts.

### 2.2 User Roles
- Platform Engineer
- SRE
- DevOps Engineer
- Auditor

### 2.3 Operating Environment
- Kubernetes v1.23+
- Multiple clusters accessible via kubeconfig
- Web UI (browser-based)

---

## 3. Functional Requirements

### 3.1 Cluster & Namespace Selection
- User SHALL select one baseline cluster + namespace
- User SHALL select one or more target cluster + namespace pairs
- Namespace names MAY differ across clusters

### 3.2 Resource Scope (Mode A)
The system SHALL support comparison of:
- Workloads: Deployment, StatefulSet, DaemonSet
- Configuration: ConfigMap, Secret
- Networking: Service

### 3.3 Comparison Rules
- Resources are matched by kind + name
- Baseline defines expected configuration
- Comparison outcomes:
  - Match
  - Mismatch
  - Not Found
  - Not Config (optional)

### 3.4 Normalization
The system SHALL normalize resources before comparison:
- Remove metadata runtime fields
- Ignore status fields
- Apply semantic sorting to known arrays
- Apply per-kind ignore rules

### 3.5 Secret Handling
- Default: compare key names only
- Optional: compare hashed values
- Secret values SHALL NOT be displayed in plain text

### 3.6 Diff Visualization
- Clicking a status cell SHALL display a normalized YAML diff
- Diff SHALL highlight changed paths
- Copy/export diff SHALL be supported

---

## 4. User Interface Requirements

### 4.1 Table Layout
- One table per resource type
- Rows: Resource names
- Columns: Target namespaces
- Cells: Status badges (Match/Mismatch/etc.)

### 4.2 Filters & Search
- Filter by status
- Search by resource name

---

## 5. Configuration Requirements

### 5.1 YAML Configuration
The system SHALL be configurable via a YAML file defining:
- Kubernetes clusters and contexts
- Resource types
- Ignore rules
- Secret comparison mode
- UI behavior

### 5.2 Configuration Validation
- Invalid configuration SHALL be rejected with clear error messages

---

## 6. Non-Functional Requirements

### 6.1 Performance
- Initial comparison SHOULD complete within acceptable time for namespaces <500 resources
- Diff view SHALL be instantaneous after initial load

### 6.2 Security
- No secret values exposed
- Read-only Kubernetes access
- Audit-safe logging

### 6.3 Scalability
- Support at least 10 clusters
- Support parallel namespace comparisons

---

## 7. Future Enhancements

### 7.1 Mode B – Policy Validation
- Rule-based assertions (OPA/Kyverno-style)
- Severity levels
- Compliance reporting

### 7.2 Export & Automation
- Export reports to JSON/HTML/Excel
- CI/CD integration
- Scheduled audits

---

## 8. Assumptions & Constraints
- Clusters are reachable via kubeconfig
- User has sufficient RBAC read permissions
- Namespaces contain logically equivalent workloads

---

## 9. Appendix
- Sample YAML configuration
- Example diff output
- Using KinD (Kubernetes in Docker) or Docker desktop for testing