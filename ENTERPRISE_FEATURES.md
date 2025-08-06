# Enterprise Features - Deliberately Not Included

This document explains features that are **intentionally not implemented** in the open-source GitHub Actions Bridge because they are provided by the ConfigHub SaaS platform.

## Philosophy

The GitHub Actions Bridge follows a clear separation:
- **Open Source**: Core workflow execution and local testing
- **ConfigHub SaaS**: Enterprise features, monitoring, and compliance

This approach avoids reinventing enterprise infrastructure that ConfigHub already provides.

## Features Delegated to ConfigHub SaaS

### 1. Metrics and Monitoring
**Not Implemented:**
- Prometheus metrics endpoint
- Performance counters
- Resource utilization tracking
- SLA/SLO monitoring

**Why:** ConfigHub provides comprehensive monitoring with:
- Built-in Prometheus/Grafana dashboards
- Workflow execution metrics
- Resource usage analytics
- Custom alerting rules

**Current State:** Basic health check endpoint only (`/health`)

### 2. Audit Logging and Compliance
**Not Implemented:**
- Structured audit logs
- User action tracking
- Compliance report generation
- Data retention policies
- Tamper-evident logging

**Why:** ConfigHub provides:
- Enterprise audit trail
- SOC 2 compliance features
- GDPR data management
- Automated compliance reporting

**Current State:** Basic operational logging for debugging

### 3. High Availability and Disaster Recovery
**Not Implemented:**
- Multi-instance coordination
- State replication
- Automatic failover
- Backup/restore procedures
- Disaster recovery orchestration

**Why:** ConfigHub provides:
- Managed HA deployment
- Automated backups
- Cross-region replication
- Zero-downtime updates

**Current State:** Single-instance deployment only

### 4. Advanced Secret Management
**Not Implemented:**
- Secret rotation automation
- Hardware Security Module (HSM) integration
- Key lifecycle management
- Secret access policies
- Encryption key management

**Why:** ConfigHub provides:
- Enterprise secret vault
- Automated rotation
- Access control policies
- Audit trail for secret access

**Current State:** Basic secret encryption and injection

### 5. Workflow Orchestration
**Not Implemented:**
- Cross-workflow dependencies
- Workflow scheduling
- Approval workflows
- Resource queueing
- Priority management

**Why:** ConfigHub provides:
- Advanced workflow orchestration
- Dependency management
- Approval gates
- Resource optimization

**Current State:** Simple sequential execution

### 6. Multi-Tenancy and RBAC
**Not Implemented:**
- User authentication
- Role-based access control
- Workspace isolation
- Resource quotas
- Team management

**Why:** ConfigHub provides:
- Enterprise SSO integration
- Fine-grained permissions
- Team collaboration
- Resource limits per tenant

**Current State:** No authentication (local use only)

## Integration with ConfigHub

To access these enterprise features:

1. **Sign up for ConfigHub**: https://confighub.com
2. **Deploy the bridge as a worker**:
   ```bash
   cub worker create github-actions-bridge
   eval "$(cub worker get-envs github-actions-bridge)"
   ./bin/actions-bridge
   ```
3. **Enterprise features activate automatically** when connected to ConfigHub

## Why This Approach?

1. **Focus**: Keep the open-source project focused on core functionality
2. **Avoid Duplication**: Don't rebuild what ConfigHub already provides
3. **Maintenance**: Reduce complexity and maintenance burden
4. **Best Practices**: Leverage enterprise-grade infrastructure

## For Enterprise Users

If you need these features:
- **Development/Testing**: Use the open-source bridge locally
- **Production**: Deploy with ConfigHub for enterprise features
- **Hybrid**: Develop locally, deploy to ConfigHub

## Related Documentation

- [README.md](README.md) - Project overview
- [SECURITY.md](SECURITY.md) - Security considerations
- [SDK_REQUESTS.md](SDK_REQUESTS.md) - Feature requests for ConfigHub SDK
- [Part 3 Planning](README.md#part-3-enterprise-features-via-confighub-saas) - Future integration plans

## Contributing

If you're interested in contributing enterprise features:
1. Consider if it belongs in ConfigHub SaaS instead
2. Discuss in GitHub Issues first
3. Keep the separation of concerns clear

Remember: Not every feature belongs in the open-source project. Sometimes the best contribution is recognizing when to use the right tool for the job.