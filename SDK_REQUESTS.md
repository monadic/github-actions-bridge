# SDK Feature Requests for ConfigHub Team

This document outlines feature requests and improvements for the [ConfigHub SDK](https://github.com/confighub/sdk) based on our experience building the GitHub Actions Bridge.

## High Priority Requests

### 1. Tagged Releases
**Current:** Using commit hashes (e.g., `v0.0.0-20250804044729-f1517379cea0`)
**Request:** Semantic versioning with tagged releases
**Benefits:**
- Clearer dependency management
- Better change tracking
- Improved stability

**Suggested Implementation:**
```bash
git tag -a v0.1.0 -m "Initial stable release"
git push origin v0.1.0
```

### 2. Structured Logging Interface
**Request:** Standardized logging interface for bridge workers
```go
type WorkerLogger interface {
    Debug(format string, args ...interface{})
    Info(format string, args ...interface{})
    Warn(format string, args ...interface{})
    Error(format string, args ...interface{})
    WithField(key string, value interface{}) WorkerLogger
}
```
**Benefits:**
- Consistent logging across all bridge implementations
- Easier integration with ConfigHub's centralized logging

### 3. Metrics Collection Interface
**Request:** Built-in metrics support for workers
```go
type WorkerMetrics interface {
    IncrementCounter(name string, labels map[string]string)
    RecordDuration(name string, duration time.Duration, labels map[string]string)
    SetGauge(name string, value float64, labels map[string]string)
}
```
**Benefits:**
- Standardized metrics across bridges
- Automatic integration with ConfigHub monitoring

### 4. Worker Health Check Standard
**Request:** Standardized health check interface
```go
type HealthChecker interface {
    CheckHealth() HealthStatus
}

type HealthStatus struct {
    Healthy bool
    Checks  map[string]CheckResult
    Version string
}
```

## Medium Priority Requests

### 5. Configuration Validation
**Request:** Built-in validation for bridge configurations
```go
type ConfigValidator interface {
    ValidateConfig(data []byte) ([]ValidationError, error)
    GetSchema() json.RawMessage
}
```

### 6. Testing Utilities
**Request:** Test helpers for bridge development
```go
// Mock worker context for testing
func NewMockWorkerContext() api.BridgeWorkerContext

// Test payload generator
func GenerateTestPayload(opts TestPayloadOptions) api.BridgeWorkerPayload
```

### 7. Example Bridge Implementation
**Request:** Complete example bridge as reference
- Full implementation of all required methods
- Best practices demonstrated
- Common patterns documented

## Low Priority Requests

### 8. CLI Scaffolding Tool
**Request:** Tool to generate bridge boilerplate
```bash
confighub-sdk scaffold bridge --name my-bridge --type terraform
```

### 9. Performance Profiling Hooks
**Request:** Optional profiling integration
```go
type ProfilerHooks interface {
    StartProfile(name string) func()
    RecordMemory(label string)
}
```

### 10. Migration Guides
**Request:** Documentation for migrating from:
- Direct API integration to SDK
- Other worker protocols to ConfigHub

## Documentation Requests

### 11. API Stability Guarantees
**Request:** Clear documentation on:
- Which interfaces are stable
- Deprecation policy
- Backward compatibility promises

### 12. Bridge Certification Program
**Request:** Process for:
- Testing bridge compliance
- Performance benchmarks
- Security review checklist

## Implementation Suggestions

### For ConfigHub SDK Team

1. **Prioritize Semantic Versioning** - This blocks production adoption
2. **Add Integration Tests** - Show real bridge/worker interaction
3. **Publish Roadmap** - Help bridge developers plan
4. **Create Bridge Registry** - Central listing of available bridges

### For Bridge Developers

While waiting for SDK improvements:
1. Use vendoring for stability
2. Implement own logging abstraction
3. Prepare for metrics interface
4. Follow example patterns in this project

## How to Submit Feedback

1. **GitHub Issues**: https://github.com/confighub/sdk/issues
2. **Pull Requests**: Contribute improvements directly
3. **Community Forum**: (if available)

## Related Documentation

- [SDK_VALIDATION.md](SDK_VALIDATION.md) - Current SDK analysis
- [ENTERPRISE_FEATURES.md](ENTERPRISE_FEATURES.md) - Features provided by ConfigHub
- [README.md](README.md) - Project overview
- [ConfigHub SDK Repo](https://github.com/confighub/sdk) - SDK source

## Acknowledgments

We appreciate the ConfigHub team's work on the SDK and look forward to collaborating on these improvements. The SDK provides a solid foundation for building bridge workers.