# Todo Tool Implementation Guide

This guide provides technical implementation details for Claude Code to apply the todo tool selection criteria, including automatic selection logic, user overrides, fallback strategies, and optimization techniques.

## Table of Contents

1. [Implementation Architecture](#implementation-architecture)
2. [Automatic Selection Logic](#automatic-selection-logic)
3. [User Override Handling](#user-override-handling)
4. [Fallback Strategies](#fallback-strategies)
5. [Performance Optimization](#performance-optimization)
6. [Integration Patterns](#integration-patterns)
7. [Monitoring and Improvement](#monitoring-and-improvement)

## Implementation Architecture

### System Components

```
┌─────────────────────────────────────────┐
│          User Request Parser            │
├─────────────────────────────────────────┤
│        Selection Criteria Engine        │
├──────────────┬──────────────────────────┤
│   Scorer     │    Decision Maker        │
├──────────────┴──────────────────────────┤
│         Tool Router                     │
├──────────────┬──────────────────────────┤
│ Native Tools │   MCP Server             │
└──────────────┴──────────────────────────┘
```

### Core Classes

```typescript
interface TodoToolSelector {
  selectTool(request: UserRequest): ToolSelection;
  scoreRequest(request: UserRequest): ScoringResult;
  applyUserOverride(request: UserRequest, selection: ToolSelection): ToolSelection;
}

interface ScoringResult {
  scores: Map<Criterion, number>;
  weightedTotal: number;
  recommendation: 'native' | 'mcp';
  confidence: number;
}

interface ToolSelection {
  tool: 'native' | 'mcp';
  reason: string;
  score?: number;
  override?: boolean;
}
```

## Automatic Selection Logic

### Request Analysis

```typescript
class RequestAnalyzer {
  analyzeRequest(request: string): RequestFeatures {
    const features: RequestFeatures = {
      // Detect multi-session indicators
      hasLongTermKeywords: /project|implement|develop|architecture|migration/i.test(request),
      
      // Detect TDD/testing keywords
      hasTddKeywords: /TDD|test.driven|RGRC|red.green|test.first/i.test(request),
      
      // Detect feature requirements
      needsSearch: /search|find|which.*todo|locate/i.test(request),
      needsTemplate: /template|boilerplate|standard/i.test(request),
      needsAnalytics: /stats|analytics|progress|report/i.test(request),
      
      // Detect complexity
      hasMultiPhase: /phase|milestone|sprint|epic/i.test(request),
      hasHierarchy: /parent|child|subtask|depends/i.test(request),
      
      // Estimate task count
      estimatedTodos: this.estimateTodoCount(request),
      
      // Detect urgency
      isUrgent: /urgent|asap|critical|emergency|hotfix/i.test(request),
      
      // User preference
      explicitToolRequest: this.detectToolPreference(request)
    };
    
    return features;
  }
  
  private estimateTodoCount(request: string): number {
    // Count task indicators
    const andCount = (request.match(/\band\b/gi) || []).length;
    const bulletPoints = (request.match(/^[\s]*[-*•]/gm) || []).length;
    const numberedItems = (request.match(/^\s*\d+\./gm) || []).length;
    
    return Math.max(1, andCount + bulletPoints + numberedItems);
  }
  
  private detectToolPreference(request: string): 'native' | 'mcp' | null {
    if (/native|simple|quick|temporary/i.test(request)) return 'native';
    if (/mcp|persistent|file|permanent/i.test(request)) return 'mcp';
    return null;
  }
}
```

### Scoring Implementation

```typescript
class CriteriaScorer {
  private weights = {
    sessionPersistence: 0.25,
    taskComplexity: 0.20,
    featureRequirements: 0.20,
    dataVolume: 0.15,
    performanceNeeds: 0.10,
    integrationRequirements: 0.10
  };
  
  scoreRequest(features: RequestFeatures): ScoringResult {
    const scores = new Map<Criterion, number>();
    
    // Session Persistence Score
    scores.set('sessionPersistence', this.scoreSessionPersistence(features));
    
    // Task Complexity Score
    scores.set('taskComplexity', this.scoreTaskComplexity(features));
    
    // Feature Requirements Score
    scores.set('featureRequirements', this.scoreFeatureRequirements(features));
    
    // Data Volume Score
    scores.set('dataVolume', this.scoreDataVolume(features));
    
    // Performance Needs Score
    scores.set('performanceNeeds', this.scorePerformanceNeeds(features));
    
    // Integration Requirements Score
    scores.set('integrationRequirements', this.scoreIntegrationRequirements(features));
    
    // Calculate weighted total
    let weightedTotal = 0;
    for (const [criterion, score] of scores) {
      weightedTotal += score * this.weights[criterion] * 10;
    }
    
    return {
      scores,
      weightedTotal,
      recommendation: weightedTotal >= 40 ? 'mcp' : 'native',
      confidence: this.calculateConfidence(weightedTotal)
    };
  }
  
  private scoreSessionPersistence(features: RequestFeatures): number {
    if (features.hasLongTermKeywords) return 8;
    if (features.hasMultiPhase) return 9;
    if (features.hasTddKeywords) return 7;
    if (features.estimatedTodos > 10) return 6;
    if (features.isUrgent) return 2;
    return 3;
  }
  
  private scoreTaskComplexity(features: RequestFeatures): number {
    let score = 2; // Base score
    
    if (features.hasTddKeywords) score += 5;
    if (features.hasMultiPhase) score += 3;
    if (features.hasHierarchy) score += 2;
    if (features.estimatedTodos > 5) score += 1;
    
    return Math.min(10, score);
  }
  
  private scoreFeatureRequirements(features: RequestFeatures): number {
    let points = 0;
    
    if (features.needsSearch) points += 3;
    if (features.needsTemplate) points += 2;
    if (features.needsAnalytics) points += 2;
    if (features.hasHierarchy) points += 2;
    if (features.hasTddKeywords) points += 1;
    
    // Convert points to score
    if (points >= 8) return 10;
    if (points >= 5) return 7;
    if (points >= 3) return 5;
    return points;
  }
  
  private calculateConfidence(score: number): number {
    // High confidence when far from threshold
    const distance = Math.abs(score - 40);
    if (distance > 20) return 0.9;
    if (distance > 10) return 0.7;
    return 0.5;
  }
}
```

### Decision Engine

```typescript
class TodoToolDecisionEngine {
  private analyzer = new RequestAnalyzer();
  private scorer = new CriteriaScorer();
  
  async selectTool(request: string, context?: Context): Promise<ToolSelection> {
    // Step 1: Analyze request
    const features = this.analyzer.analyzeRequest(request);
    
    // Step 2: Check for explicit user preference
    if (features.explicitToolRequest) {
      return {
        tool: features.explicitToolRequest,
        reason: `User explicitly requested ${features.explicitToolRequest} tools`,
        override: true
      };
    }
    
    // Step 3: Score request
    const scoringResult = this.scorer.scoreRequest(features);
    
    // Step 4: Check tool availability
    const mcpAvailable = await this.checkMcpAvailability();
    
    // Step 5: Make decision
    let selectedTool = scoringResult.recommendation;
    let reason = this.generateReason(scoringResult);
    
    // Step 6: Apply fallback if needed
    if (selectedTool === 'mcp' && !mcpAvailable) {
      selectedTool = 'native';
      reason = 'MCP Server unavailable, falling back to native tools';
    }
    
    // Step 7: Log decision for monitoring
    this.logDecision(request, features, scoringResult, selectedTool);
    
    return {
      tool: selectedTool,
      reason,
      score: scoringResult.weightedTotal
    };
  }
  
  private generateReason(result: ScoringResult): string {
    const score = result.weightedTotal;
    const tool = result.recommendation;
    
    if (result.confidence > 0.8) {
      return `Clear ${tool} tools fit (score: ${score.toFixed(1)}/100)`;
    } else if (result.confidence > 0.6) {
      return `Recommended ${tool} tools (score: ${score.toFixed(1)}/100)`;
    } else {
      return `Slight preference for ${tool} tools (score: ${score.toFixed(1)}/100, threshold: 40)`;
    }
  }
}
```

## User Override Handling

### Override Detection Patterns

```typescript
const OVERRIDE_PATTERNS = {
  native: [
    /use\s+native\s+tool/i,
    /quick\s+todo/i,
    /simple\s+todo/i,
    /temporary\s+todo/i,
    /session\s+todo/i,
    /don't\s+save/i,
    /no\s+persistence/i
  ],
  mcp: [
    /use\s+mcp/i,
    /persistent\s+todo/i,
    /save\s+todo/i,
    /create\s+todo\s+file/i,
    /long[\s-]term\s+todo/i,
    /searchable\s+todo/i,
    /todo\s+with\s+template/i
  ]
};

function detectOverride(request: string): 'native' | 'mcp' | null {
  for (const pattern of OVERRIDE_PATTERNS.native) {
    if (pattern.test(request)) return 'native';
  }
  
  for (const pattern of OVERRIDE_PATTERNS.mcp) {
    if (pattern.test(request)) return 'mcp';
  }
  
  return null;
}
```

### Override Confirmation

```typescript
async function confirmOverride(
  userChoice: 'native' | 'mcp',
  recommendation: 'native' | 'mcp',
  confidence: number
): Promise<boolean> {
  if (userChoice === recommendation) {
    return true; // No override needed
  }
  
  if (confidence < 0.7) {
    return true; // Low confidence, respect user choice
  }
  
  // High confidence but user disagrees
  console.log(`
Note: Based on the task characteristics, ${recommendation} tools 
were recommended (confidence: ${(confidence * 100).toFixed(0)}%).
You've requested ${userChoice} tools. Proceeding with your choice.
  `);
  
  return true;
}
```

## Fallback Strategies

### MCP Server Unavailability

```typescript
class FallbackHandler {
  async handleMcpUnavailable(request: string): Promise<ToolSelection> {
    console.log("MCP Server is unavailable. Using native tools with limitations:");
    console.log("- Todos will not persist beyond this session");
    console.log("- Search and advanced features unavailable");
    console.log("- Consider saving important tasks elsewhere");
    
    return {
      tool: 'native',
      reason: 'MCP Server unavailable - using session-only native tools',
      score: 0
    };
  }
  
  async handleMcpError(error: Error, request: string): Promise<ToolSelection> {
    if (error.message.includes('timeout')) {
      // Try once more with longer timeout
      return this.retryWithTimeout();
    }
    
    if (error.message.includes('disk full')) {
      // Suggest cleanup
      console.log("Disk space issue detected. Run todo_clean to archive old todos.");
    }
    
    // Default fallback
    return this.handleMcpUnavailable(request);
  }
}
```

### Graceful Degradation

```typescript
class GracefulDegradation {
  async degradeFeatures(fullRequest: any): Promise<any> {
    // If MCP search fails, provide manual alternative
    if (fullRequest.operation === 'search') {
      console.log("Search unavailable. Recent todos:");
      return this.listRecentTodos();
    }
    
    // If templates fail, provide basic structure
    if (fullRequest.template) {
      console.log("Template unavailable. Creating basic todo structure.");
      return this.createBasicStructure(fullRequest.task);
    }
    
    // Convert complex operations to simple ones
    if (fullRequest.operation === 'update_section') {
      console.log("Section updates unavailable. Replacing entire todo.");
      return this.convertToFullUpdate(fullRequest);
    }
    
    return fullRequest;
  }
}
```

## Performance Optimization

### Caching Strategy

```typescript
class TodoToolCache {
  private decisionCache = new Map<string, ToolSelection>();
  private featureCache = new Map<string, RequestFeatures>();
  private mcpStatusCache: { available: boolean; timestamp: number } | null = null;
  
  getCachedDecision(request: string): ToolSelection | null {
    // Normalize request for better cache hits
    const normalized = this.normalizeRequest(request);
    return this.decisionCache.get(normalized);
  }
  
  cacheDecision(request: string, decision: ToolSelection): void {
    const normalized = this.normalizeRequest(request);
    this.decisionCache.set(normalized, decision);
    
    // Limit cache size
    if (this.decisionCache.size > 100) {
      const firstKey = this.decisionCache.keys().next().value;
      this.decisionCache.delete(firstKey);
    }
  }
  
  async getMcpStatus(): Promise<boolean> {
    // Cache MCP availability for 5 minutes
    if (this.mcpStatusCache && 
        Date.now() - this.mcpStatusCache.timestamp < 300000) {
      return this.mcpStatusCache.available;
    }
    
    const available = await this.checkMcpAvailability();
    this.mcpStatusCache = { available, timestamp: Date.now() };
    return available;
  }
  
  private normalizeRequest(request: string): string {
    return request
      .toLowerCase()
      .replace(/\s+/g, ' ')
      .trim();
  }
}
```

### Batch Operations

```typescript
class BatchOperationOptimizer {
  private pendingDecisions: Array<{
    request: string;
    resolver: (selection: ToolSelection) => void;
  }> = [];
  
  async queueDecision(request: string): Promise<ToolSelection> {
    return new Promise((resolve) => {
      this.pendingDecisions.push({ request, resolver: resolve });
      
      // Process batch after small delay
      if (this.pendingDecisions.length === 1) {
        setTimeout(() => this.processBatch(), 10);
      }
    });
  }
  
  private async processBatch(): Promise<void> {
    const batch = [...this.pendingDecisions];
    this.pendingDecisions = [];
    
    // Analyze all requests together for better context
    const decisions = await this.batchAnalyze(batch.map(b => b.request));
    
    // Resolve all promises
    batch.forEach((item, index) => {
      item.resolver(decisions[index]);
    });
  }
  
  private async batchAnalyze(requests: string[]): Promise<ToolSelection[]> {
    // Share analysis work across similar requests
    const sharedFeatures = this.extractSharedFeatures(requests);
    
    return requests.map(request => 
      this.decideWithSharedContext(request, sharedFeatures)
    );
  }
}
```

### Preemptive Loading

```typescript
class PreemptiveLoader {
  async warmupTools(context: Context): Promise<void> {
    // Based on context, preload likely tools
    if (context.previousTool === 'mcp') {
      // Likely to continue with MCP
      await this.warmupMcp();
    }
    
    if (context.timeOfDay === 'morning') {
      // Morning often has planning tasks
      await this.warmupNative();
    }
    
    if (context.projectType === 'large') {
      // Large projects need MCP
      await this.warmupMcp();
      await this.preloadTemplates();
    }
  }
  
  private async warmupMcp(): Promise<void> {
    // Pre-check MCP availability
    // Pre-load common templates
    // Pre-initialize search index
  }
}
```

## Integration Patterns

### Seamless Tool Switching

```typescript
class TodoToolSwitcher {
  async switchTools(
    fromTool: 'native' | 'mcp',
    toTool: 'native' | 'mcp',
    todos: any[]
  ): Promise<void> {
    if (fromTool === toTool) return;
    
    console.log(`Switching from ${fromTool} to ${toTool} tools...`);
    
    if (fromTool === 'native' && toTool === 'mcp') {
      // Migrate to persistent storage
      for (const todo of todos) {
        await todo_create({
          task: todo.content,
          priority: todo.priority,
          metadata: { migrated_from: 'native', original_id: todo.id }
        });
      }
      console.log(`Migrated ${todos.length} todos to persistent storage`);
    } else {
      // Extract essential info for native tools
      const simplified = todos.map(t => ({
        id: t.todo_id || t.id,
        content: t.task || t.content,
        status: t.status,
        priority: t.priority
      }));
      
      await TodoWrite({ todos: simplified });
      console.log(`Loaded ${todos.length} todos into session memory`);
    }
  }
}
```

### Unified Interface

```typescript
class UnifiedTodoInterface {
  private toolSelector = new TodoToolDecisionEngine();
  private nativeTools = new NativeTodoTools();
  private mcpTools = new McpTodoTools();
  
  async createTodo(request: string, options?: any): Promise<any> {
    // Select appropriate tool
    const selection = await this.toolSelector.selectTool(request);
    
    // Route to selected tool
    if (selection.tool === 'native') {
      return this.nativeTools.create(request, options);
    } else {
      return this.mcpTools.create(request, options);
    }
  }
  
  async searchTodos(query: string): Promise<any> {
    // Search always tries MCP first, falls back to listing
    try {
      return await this.mcpTools.search(query);
    } catch (error) {
      console.log("Search unavailable, listing all todos instead");
      return this.nativeTools.readAll();
    }
  }
}
```

## Monitoring and Improvement

### Decision Logging

```typescript
interface DecisionLog {
  timestamp: Date;
  request: string;
  features: RequestFeatures;
  scores: Map<string, number>;
  selection: 'native' | 'mcp';
  userOverride?: boolean;
  actualUsage?: {
    tool: 'native' | 'mcp';
    duration: number;
    operations: string[];
  };
}

class DecisionMonitor {
  private logs: DecisionLog[] = [];
  
  logDecision(log: DecisionLog): void {
    this.logs.push(log);
    
    // Analyze patterns periodically
    if (this.logs.length % 100 === 0) {
      this.analyzePatterns();
    }
  }
  
  private analyzePatterns(): void {
    // Calculate accuracy
    const correctDecisions = this.logs.filter(log => 
      !log.userOverride && log.actualUsage?.tool === log.selection
    ).length;
    
    const accuracy = correctDecisions / this.logs.length;
    
    // Identify problem areas
    const overrides = this.logs.filter(log => log.userOverride);
    const commonOverridePatterns = this.extractPatterns(overrides);
    
    // Log insights
    console.log(`Decision accuracy: ${(accuracy * 100).toFixed(1)}%`);
    console.log(`Common override patterns:`, commonOverridePatterns);
  }
}
```

### Adaptive Scoring

```typescript
class AdaptiveScorer {
  private scoreAdjustments = new Map<string, number>();
  
  async updateScoringWeights(feedback: DecisionLog[]): Promise<void> {
    // Analyze where scoring was wrong
    const mismatches = feedback.filter(log => 
      log.actualUsage?.tool !== log.selection && !log.userOverride
    );
    
    for (const mismatch of mismatches) {
      // Identify which criteria led to wrong decision
      const problematicCriteria = this.identifyProblematicCriteria(mismatch);
      
      // Adjust weights slightly
      for (const criterion of problematicCriteria) {
        const currentAdjustment = this.scoreAdjustments.get(criterion) || 0;
        const newAdjustment = currentAdjustment + (mismatch.selection === 'mcp' ? -0.01 : 0.01);
        this.scoreAdjustments.set(criterion, newAdjustment);
      }
    }
  }
  
  getAdjustedWeight(criterion: string, baseWeight: number): number {
    const adjustment = this.scoreAdjustments.get(criterion) || 0;
    return Math.max(0.05, Math.min(0.40, baseWeight + adjustment));
  }
}
```

### Performance Metrics

```typescript
class PerformanceTracker {
  private metrics = {
    decisionTime: [] as number[],
    toolSwitches: 0,
    cacheHits: 0,
    cacheMisses: 0,
    fallbacks: 0
  };
  
  recordDecisionTime(start: number): void {
    const duration = Date.now() - start;
    this.metrics.decisionTime.push(duration);
  }
  
  getPerformanceReport(): string {
    const avgDecisionTime = this.metrics.decisionTime.reduce((a, b) => a + b, 0) / 
                           this.metrics.decisionTime.length;
    
    const cacheHitRate = this.metrics.cacheHits / 
                        (this.metrics.cacheHits + this.metrics.cacheMisses);
    
    return `
Performance Report:
- Avg Decision Time: ${avgDecisionTime.toFixed(2)}ms
- Cache Hit Rate: ${(cacheHitRate * 100).toFixed(1)}%
- Tool Switches: ${this.metrics.toolSwitches}
- Fallbacks Used: ${this.metrics.fallbacks}
    `;
  }
}
```

## Implementation Checklist

### Phase 1: Basic Selection (Week 1)
- [ ] Implement RequestAnalyzer
- [ ] Implement CriteriaScorer  
- [ ] Create TodoToolDecisionEngine
- [ ] Add basic override detection
- [ ] Test with common scenarios

### Phase 2: Optimization (Week 2)
- [ ] Add decision caching
- [ ] Implement batch analysis
- [ ] Add MCP availability caching
- [ ] Create performance tracker
- [ ] Optimize request parsing

### Phase 3: Advanced Features (Week 3)
- [ ] Implement adaptive scoring
- [ ] Add decision monitoring
- [ ] Create unified interface
- [ ] Add graceful degradation
- [ ] Implement tool switching

### Phase 4: Production Ready (Week 4)
- [ ] Complete test coverage
- [ ] Add comprehensive logging
- [ ] Document all patterns
- [ ] Performance benchmarks
- [ ] User feedback integration

## Conclusion

This implementation guide provides a robust framework for intelligent todo tool selection that:

1. **Automatically selects** the best tool based on task characteristics
2. **Respects user preferences** through override detection
3. **Handles failures gracefully** with fallback strategies
4. **Optimizes performance** through caching and batching
5. **Improves over time** through monitoring and adaptation

The system balances automation with user control, ensuring that the right tool is used for each task while maintaining flexibility for user preferences and edge cases.