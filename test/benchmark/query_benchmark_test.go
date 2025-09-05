package benchmark

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
)

const (
	maxQueryTime     = 50 * time.Millisecond
	maxMemoryUsageMB = 50
	numAgentsTarget  = 1000
)

func BenchmarkQueryPerformance(b *testing.B) {
	testDir := createBenchmarkTestDir(b)
	defer os.RemoveAll(testDir)

	generateTestAgents(b, testDir, numAgentsTarget)
	queryEngine := setupQueryEngine(b, testDir)

	queries := []string{
		"golang",
		"python OR javascript",
		"web AND framework",
		"name:general*",
		"description:\"API integration\"",
	}

	b.Run("QueryLatency", func(b *testing.B) {
		for _, query := range queries {
			b.Run(fmt.Sprintf("Query_%s", sanitizeQueryName(query)), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					start := time.Now()
					_, err := queryEngine.Query(query, engine.QueryOptions{Context: context.Background()})
					elapsed := time.Since(start)

					if err != nil {
						b.Fatalf("Query failed: %v", err)
					}

					if elapsed > maxQueryTime {
						b.Errorf("Query took %v, expected <%v", elapsed, maxQueryTime)
					}
				}
			})
		}
	})

	b.Run("ConcurrentQueries", func(b *testing.B) {
		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			queryIndex := 0
			for pb.Next() {
				query := queries[queryIndex%len(queries)]
				queryIndex++

				start := time.Now()
				_, err := queryEngine.Query(query, engine.QueryOptions{Context: context.Background()})
				elapsed := time.Since(start)

				if err != nil {
					b.Errorf("Concurrent query failed: %v", err)
				}

				if elapsed > maxQueryTime {
					b.Errorf("Concurrent query took %v, expected <%v", elapsed, maxQueryTime)
				}
			}
		})
	})
}

func BenchmarkMemoryUsage(b *testing.B) {
	testDir := createBenchmarkTestDir(b)
	defer os.RemoveAll(testDir)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	generateTestAgents(b, testDir, numAgentsTarget)
	queryEngine := setupQueryEngine(b, testDir)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memoryUsedMB := float64(m2.Alloc-m1.Alloc) / (1024 * 1024)

	b.ReportMetric(memoryUsedMB, "MB")

	if memoryUsedMB > maxMemoryUsageMB {
		b.Errorf("Memory usage %.2f MB exceeds limit of %d MB", memoryUsedMB, maxMemoryUsageMB)
	}

	queries := []string{"golang", "python", "web framework"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, query := range queries {
			_, err := queryEngine.Query(query, engine.QueryOptions{Context: context.Background()})
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}
		}
	}
}

func BenchmarkEngineCreation(b *testing.B) {
	testDir := createBenchmarkTestDir(b)
	defer os.RemoveAll(testDir)

	generateTestAgents(b, testDir, numAgentsTarget)

	b.Run("EngineCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engineDir := filepath.Join(testDir, fmt.Sprintf("engine_%d", i))
			indexPath := filepath.Join(engineDir, "index")
			cachePath := filepath.Join(engineDir, "cache")

			start := time.Now()
			queryEngine, err := engine.NewEngine(indexPath, cachePath)
			elapsed := time.Since(start)

			if err != nil {
				b.Fatalf("Engine creation failed: %v", err)
			}

			b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/creation")
			_ = queryEngine
			os.RemoveAll(engineDir)
		}
	})
}

func BenchmarkCachePerformance(b *testing.B) {
	testDir := createBenchmarkTestDir(b)
	defer os.RemoveAll(testDir)

	generateTestAgents(b, testDir, 100)
	queryEngine := setupQueryEngine(b, testDir)

	testQueries := make([]string, 100)
	for i := 0; i < 100; i++ {
		testQueries[i] = fmt.Sprintf("test_query_%d", i)
	}

	b.Run("CacheWrite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			query := testQueries[i%len(testQueries)]
			_, _ = queryEngine.Query(query, engine.QueryOptions{Context: context.Background()})
		}
	})

	for _, query := range testQueries {
		_, _ = queryEngine.Query(query, engine.QueryOptions{Context: context.Background()})
	}

	b.Run("CacheRead", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			query := testQueries[i%len(testQueries)]
			_, _ = queryEngine.Query(query, engine.QueryOptions{Context: context.Background()})
		}
	})
}

func createBenchmarkTestDir(b *testing.B) string {
	testDir, err := os.MkdirTemp("", "agent_benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}
	return testDir
}

func generateTestAgents(b *testing.B, testDir string, count int) {
	agentsDir := filepath.Join(testDir, "agents")
	err := os.MkdirAll(agentsDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create agents directory: %v", err)
	}

	agentTypes := []string{"general-purpose", "code-reviewer", "go-specialist", "python-expert", "web-developer"}
	frameworks := []string{"react", "vue", "django", "flask", "gin", "echo", "express"}

	for i := 0; i < count; i++ {
		agentType := agentTypes[i%len(agentTypes)]
		framework := frameworks[i%len(frameworks)]

		agentContent := fmt.Sprintf(`---
name: test-agent-%d
type: %s
description: A test agent for %s development
tools: ["editor", "bash", "git"]
---

# Test Agent %d

This is a test agent used for benchmarking the query system.

## Tools Available

- Development tools for %s
- %s integration
- Performance optimization
- Code review capabilities

## Usage

This agent specializes in %s development and provides comprehensive support for %s projects.

## Examples

bash
# Example command
test-command --agent=%d --type=%s

## Performance Notes

This agent is designed to handle high-throughput scenarios efficiently.
`, i, agentType, framework, i, framework, framework, framework, framework, i, agentType)

		agentFile := filepath.Join(agentsDir, fmt.Sprintf("agent-%d.md", i))
		err := os.WriteFile(agentFile, []byte(agentContent), 0644)
		if err != nil {
			b.Fatalf("Failed to write agent file: %v", err)
		}
	}
}

func setupQueryEngine(b *testing.B, testDir string) *engine.Engine {
	indexPath := filepath.Join(testDir, "index")
	cachePath := filepath.Join(testDir, "cache")

	queryEngine, err := engine.NewEngine(indexPath, cachePath)
	if err != nil {
		b.Fatalf("Failed to create query engine: %v", err)
	}

	return queryEngine
}

func sanitizeQueryName(query string) string {
	sanitized := ""
	for _, char := range query {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			sanitized += string(char)
		} else {
			sanitized += "_"
		}
	}
	return sanitized
}
