---
template_name: prd
description: Product Requirements Document template for new features and projects
variables:
  - feature_name
  - author
  - date
---
# PRD: {{.feature_name}}

**Author**: {{.author}}  
**Date**: {{.date}}  
**Status**: Draft

## Executive Summary

*Provide a brief overview of the feature/product in 2-3 sentences. What is it? Why does it matter?*

## Problem Statement

*Describe the problem this feature solves. Include:*
- Who experiences this problem?
- What is the impact of not solving it?
- Current workarounds (if any)

## Goals & Success Metrics

### Primary Goals
1. *Main objective*
2. *Secondary objective*

### Success Metrics
- **Metric 1**: *e.g., Increase user engagement by 15%*
- **Metric 2**: *e.g., Reduce task completion time by 30%*
- **Metric 3**: *e.g., Achieve 90% user satisfaction score*

## User Stories

### As a [user type]
- I want to [action]
- So that [benefit]
- **Acceptance Criteria**:
  - [ ] Criterion 1
  - [ ] Criterion 2

### As a [user type]
- I want to [action]
- So that [benefit]
- **Acceptance Criteria**:
  - [ ] Criterion 1
  - [ ] Criterion 2

## Requirements

### Functional Requirements
1. **FR1**: *Core functionality description*
2. **FR2**: *Additional feature description*
3. **FR3**: *Integration requirement*

### Non-Functional Requirements
1. **Performance**: *e.g., Response time < 100ms*
2. **Scalability**: *e.g., Support 10,000 concurrent users*
3. **Security**: *e.g., OAuth 2.0 authentication*
4. **Compatibility**: *e.g., Works with existing API*

### Out of Scope
- *Feature/functionality not included in this phase*
- *Future enhancement to consider later*

## Technical Approach

### Architecture Overview
*High-level technical design and key components*

### Technology Stack
- **Frontend**: *e.g., React, TypeScript*
- **Backend**: *e.g., Go, PostgreSQL*
- **Infrastructure**: *e.g., AWS, Docker*

### API Design
*Key endpoints and data models if applicable*

## Risks & Assumptions

### Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| *Technical risk* | High | Medium | *Mitigation strategy* |
| *Resource risk* | Medium | Low | *Mitigation strategy* |

### Assumptions
- *Assumption about users/technology/resources*
- *Dependencies on other teams/systems*

### Dependencies
- *External system or team dependency*
- *Required infrastructure or tools*

## Timeline & Milestones

### Phase 1: Foundation (Week 1-2)
- [ ] Technical design review
- [ ] Set up development environment
- [ ] Create initial prototypes

### Phase 2: Core Development (Week 3-4)
- [ ] Implement core functionality
- [ ] Write unit tests
- [ ] Internal testing

### Phase 3: Polish & Launch (Week 5-6)
- [ ] User acceptance testing
- [ ] Documentation
- [ ] Production deployment

## Open Questions

1. *Question requiring stakeholder input?*
2. *Technical decision to be made?*
3. *Resource allocation question?*

---

## Appendix

### References
- *Link to related docs*
- *Design mockups*
- *Technical specifications*

### Glossary
- **Term**: Definition
- **Acronym**: Explanation