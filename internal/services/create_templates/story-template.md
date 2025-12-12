---
id: {{.ID}}
title: {{.Title}}
status: {{.Status}}
priority: {{.Priority}}
{{- if .Assignee}}
assignee: {{.Assignee}}
{{- end}}
{{- if .CreatedAt}}
created_at: {{.CreatedAt}}
{{- end}}
{{- if .Tags}}
tags:
{{- range .Tags}}
  - {{.}}
{{- end}}
{{- end}}
---

# {{.Title}}

<!-- Add your story description, acceptance criteria, and notes here -->
