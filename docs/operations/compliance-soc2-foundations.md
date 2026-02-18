# SOC 2 Foundations Checklist

This playbook provides a practical baseline for SOC 2-oriented engineering and operations.

It is not legal advice and not a certification substitute.

## 1) Access control

- Enforce least privilege for production systems.
- Use role-based access for app and infrastructure administration.
- Require MFA for admin accounts.
- Maintain joiner/mover/leaver access procedures.

Evidence examples:

- Access review logs
- Role definitions
- Offboarding checklist records

## 2) Change management

- Require pull request review for production-bound changes.
- Require CI checks before merge to protected branches.
- Keep release history and deployment logs.
- Document emergency change procedure and post-incident review process.

Evidence examples:

- PR approvals and CI status history
- Deployment audit trail
- Incident/postmortem records

## 3) Secure development lifecycle

- Use dependency vulnerability scanning.
- Patch critical vulnerabilities within defined SLA.
- Add basic secure coding checks for auth, tenancy, and data handling changes.
- Maintain a threat/risk register for major features.

Evidence examples:

- Vulnerability reports and remediation tickets
- Risk register updates
- Security review checklists

## 4) Data protection

- Encrypt data in transit and at rest.
- Store secrets in environment/secret manager only, never in code.
- Redact sensitive values in logs and error payloads.
- Define backup and restore processes for critical data stores.

Evidence examples:

- Secret management policy
- Backup job logs and restore test records
- Logging redaction verification notes

## 5) Monitoring and incident response

- Centralize logs and alerts for critical services.
- Define incident severity levels and escalation paths.
- Track and resolve security events with ticketed workflow.
- Conduct periodic incident-response exercises.

Evidence examples:

- Alert runbooks
- Incident tickets
- Exercise notes and action items

## 6) Vendor and dependency governance

- Keep an inventory of critical vendors and services.
- Track contracts and security posture for critical vendors.
- Review vendor changes that impact authentication, billing, or customer data.

Evidence examples:

- Vendor inventory with owners
- Security review checklists
- Contract/assessment records

## 7) Documentation and training

- Keep architecture/security docs current with implementation.
- Train contributors on tenancy, auth boundary, and data-handling guardrails.
- Enforce checklist usage in PR templates or release gates.

Evidence examples:

- Training completion records
- Documentation update history
- PR checklist compliance reports

## Engineering baseline for this template

- All tenant-scoped operations enforce organization membership.
- Auth and billing provider interactions are isolated in adapters.
- Audit events are generated for identity, tenant, and billing-sensitive changes.
- Sensitive fields are excluded or redacted from logs by default.
