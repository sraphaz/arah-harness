id: {{DOMAIN_ID}}
name: {{DOMAIN_NAME}}
description: {{DOMAIN_DESCRIPTION}}
type: domain
triggers:
  - pull_request:opened
  - pull_request:synchronize
scope:
  paths:
{{DOMAIN_PATHS}}
  may_code: false
checklist: checklists/_shared.conduct.md
guardrails:
  no_merge: true
  consult_only: true
