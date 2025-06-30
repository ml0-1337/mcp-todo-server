package core

import (
	"testing"
	"time"
)

// Helper function to create a test todo
func createTestTodo(id, task, parentID string) *Todo {
	return &Todo{
		ID:       id,
		Task:     task,
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Started:  time.Now(),
		ParentID: parentID,
	}
}

// Helper function to create a test node
func createTestNode(id, task, parentID string) *TodoNode {
	return &TodoNode{
		Todo:     createTestTodo(id, task, parentID),
		Children: []*TodoNode{},
	}
}

// Helper function to build a simple hierarchy
func buildTestHierarchy() []*TodoNode {
	// Create a hierarchy:
	// root1
	//   ├── child1
	//   │   └── grandchild1
	//   └── child2
	// root2
	//   └── child3

	root1 := createTestNode("root1", "Root Task 1", "")
	child1 := createTestNode("child1", "Child Task 1", "root1")
	grandchild1 := createTestNode("grandchild1", "Grandchild Task 1", "child1")
	child2 := createTestNode("child2", "Child Task 2", "root1")
	root2 := createTestNode("root2", "Root Task 2", "")
	child3 := createTestNode("child3", "Child Task 3", "root2")

	// Build relationships
	child1.Children = append(child1.Children, grandchild1)
	root1.Children = append(root1.Children, child1, child2)
	root2.Children = append(root2.Children, child3)

	return []*TodoNode{root1, root2}
}

// TestCountHierarchyTodos tests counting todos in a hierarchy
func TestCountHierarchyTodos(t *testing.T) {
	tests := []struct {
		name     string
		roots    []*TodoNode
		expected int
	}{
		{
			name:     "empty hierarchy",
			roots:    []*TodoNode{},
			expected: 0,
		},
		{
			name: "single node",
			roots: []*TodoNode{
				createTestNode("single", "Single Task", ""),
			},
			expected: 1,
		},
		{
			name:     "complex hierarchy",
			roots:    buildTestHierarchy(),
			expected: 6, // 2 roots + 4 children
		},
		{
			name: "multiple single nodes",
			roots: []*TodoNode{
				createTestNode("node1", "Node 1", ""),
				createTestNode("node2", "Node 2", ""),
				createTestNode("node3", "Node 3", ""),
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountHierarchyTodos(tt.roots)
			if result != tt.expected {
				t.Errorf("CountHierarchyTodos() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestCountNodeTodos tests recursive node counting
func TestCountNodeTodos(t *testing.T) {
	// Build test hierarchy
	roots := buildTestHierarchy()

	tests := []struct {
		name     string
		node     *TodoNode
		expected int
	}{
		{
			name:     "root with children",
			node:     roots[0], // root1 with 2 children and 1 grandchild
			expected: 4,
		},
		{
			name:     "node with single child",
			node:     roots[0].Children[0], // child1 with grandchild1
			expected: 2,
		},
		{
			name:     "leaf node",
			node:     roots[0].Children[0].Children[0], // grandchild1
			expected: 1,
		},
		{
			name:     "root with single child",
			node:     roots[1], // root2 with child3
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countNodeTodos(tt.node)
			if result != tt.expected {
				t.Errorf("countNodeTodos() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestFindNodeByID tests finding a node by ID in the hierarchy
func TestFindNodeByID(t *testing.T) {
	roots := buildTestHierarchy()

	tests := []struct {
		name        string
		id          string
		shouldFind  bool
		expectedID  string
	}{
		{
			name:        "find root node",
			id:          "root1",
			shouldFind:  true,
			expectedID:  "root1",
		},
		{
			name:        "find child node",
			id:          "child1",
			shouldFind:  true,
			expectedID:  "child1",
		},
		{
			name:        "find grandchild node",
			id:          "grandchild1",
			shouldFind:  true,
			expectedID:  "grandchild1",
		},
		{
			name:        "find node in second tree",
			id:          "child3",
			shouldFind:  true,
			expectedID:  "child3",
		},
		{
			name:        "node not found",
			id:          "nonexistent",
			shouldFind:  false,
		},
		{
			name:        "empty ID",
			id:          "",
			shouldFind:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNodeByID(roots, tt.id)
			if tt.shouldFind {
				if result == nil {
					t.Errorf("FindNodeByID() expected to find node with ID %s, but got nil", tt.id)
				} else if result.Todo.ID != tt.expectedID {
					t.Errorf("FindNodeByID() found wrong node: got ID %s, want %s", result.Todo.ID, tt.expectedID)
				}
			} else {
				if result != nil {
					t.Errorf("FindNodeByID() expected nil, but found node with ID %s", result.Todo.ID)
				}
			}
		})
	}
}

// TestFindNodeInTree tests recursive tree search
func TestFindNodeInTree(t *testing.T) {
	roots := buildTestHierarchy()

	tests := []struct {
		name       string
		node       *TodoNode
		id         string
		shouldFind bool
	}{
		{
			name:       "find in same node",
			node:       roots[0],
			id:         "root1",
			shouldFind: true,
		},
		{
			name:       "find in children",
			node:       roots[0],
			id:         "child1",
			shouldFind: true,
		},
		{
			name:       "find in grandchildren",
			node:       roots[0],
			id:         "grandchild1",
			shouldFind: true,
		},
		{
			name:       "not in subtree",
			node:       roots[0],
			id:         "child3", // This is in root2's tree
			shouldFind: false,
		},
		{
			name:       "find in leaf's subtree",
			node:       roots[0].Children[0].Children[0], // grandchild1
			id:         "grandchild1",
			shouldFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findNodeInTree(tt.node, tt.id)
			if tt.shouldFind && result == nil {
				t.Errorf("findNodeInTree() expected to find node with ID %s", tt.id)
			} else if !tt.shouldFind && result != nil {
				t.Errorf("findNodeInTree() expected nil, but found node")
			}
		})
	}
}

// TestGetNodePath tests getting the path from root to a specific node
func TestGetNodePath(t *testing.T) {
	roots := buildTestHierarchy()

	tests := []struct {
		name         string
		id           string
		expectedPath []string // IDs in the path
	}{
		{
			name:         "path to root",
			id:           "root1",
			expectedPath: []string{"root1"},
		},
		{
			name:         "path to child",
			id:           "child1",
			expectedPath: []string{"root1", "child1"},
		},
		{
			name:         "path to grandchild",
			id:           "grandchild1",
			expectedPath: []string{"root1", "child1", "grandchild1"},
		},
		{
			name:         "path in second tree",
			id:           "child3",
			expectedPath: []string{"root2", "child3"},
		},
		{
			name:         "node not found",
			id:           "nonexistent",
			expectedPath: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNodePath(roots, tt.id)
			
			if tt.expectedPath == nil {
				if result != nil {
					t.Errorf("GetNodePath() expected nil, but got path")
				}
				return
			}

			if len(result) != len(tt.expectedPath) {
				t.Errorf("GetNodePath() path length = %d, want %d", len(result), len(tt.expectedPath))
				return
			}

			for i, node := range result {
				if node.Todo.ID != tt.expectedPath[i] {
					t.Errorf("GetNodePath()[%d] = %s, want %s", i, node.Todo.ID, tt.expectedPath[i])
				}
			}
		})
	}
}

// TestGetPathToNode tests recursive path building
func TestGetPathToNode(t *testing.T) {
	roots := buildTestHierarchy()

	tests := []struct {
		name         string
		node         *TodoNode
		targetID     string
		expectedPath []string
	}{
		{
			name:         "path to self",
			node:         roots[0],
			targetID:     "root1",
			expectedPath: []string{"root1"},
		},
		{
			name:         "path to child",
			node:         roots[0],
			targetID:     "child1",
			expectedPath: []string{"root1", "child1"},
		},
		{
			name:         "path to grandchild",
			node:         roots[0],
			targetID:     "grandchild1",
			expectedPath: []string{"root1", "child1", "grandchild1"},
		},
		{
			name:         "target not in subtree",
			node:         roots[0],
			targetID:     "child3",
			expectedPath: nil,
		},
		{
			name:         "path from middle node",
			node:         roots[0].Children[0], // child1
			targetID:     "grandchild1",
			expectedPath: []string{"child1", "grandchild1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPathToNode(tt.node, tt.targetID, []*TodoNode{})
			
			if tt.expectedPath == nil {
				if result != nil {
					t.Errorf("getPathToNode() expected nil, but got path")
				}
				return
			}

			if len(result) != len(tt.expectedPath) {
				t.Errorf("getPathToNode() path length = %d, want %d", len(result), len(tt.expectedPath))
				return
			}

			for i, node := range result {
				if node.Todo.ID != tt.expectedPath[i] {
					t.Errorf("getPathToNode()[%d] = %s, want %s", i, node.Todo.ID, tt.expectedPath[i])
				}
			}
		})
	}
}

// TestFlattenHierarchy tests converting hierarchy to flat list
func TestFlattenHierarchy(t *testing.T) {
	tests := []struct {
		name         string
		roots        []*TodoNode
		expectedIDs  []string
	}{
		{
			name:         "empty hierarchy",
			roots:        []*TodoNode{},
			expectedIDs:  []string{},
		},
		{
			name: "single node",
			roots: []*TodoNode{
				createTestNode("single", "Single Task", ""),
			},
			expectedIDs: []string{"single"},
		},
		{
			name:         "complex hierarchy",
			roots:        buildTestHierarchy(),
			expectedIDs:  []string{"root1", "child1", "grandchild1", "child2", "root2", "child3"},
		},
		{
			name: "wide hierarchy",
			roots: func() []*TodoNode {
				root := createTestNode("root", "Root", "")
				for i := 1; i <= 5; i++ {
					child := createTestNode(string(rune('0'+i)), "Child", "root")
					root.Children = append(root.Children, child)
				}
				return []*TodoNode{root}
			}(),
			expectedIDs: []string{"root", "1", "2", "3", "4", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenHierarchy(tt.roots)
			
			if len(result) != len(tt.expectedIDs) {
				t.Errorf("FlattenHierarchy() returned %d todos, want %d", len(result), len(tt.expectedIDs))
				return
			}

			for i, todo := range result {
				if todo.ID != tt.expectedIDs[i] {
					t.Errorf("FlattenHierarchy()[%d].ID = %s, want %s", i, todo.ID, tt.expectedIDs[i])
				}
			}
		})
	}
}

// TestFlattenNode tests recursive node flattening
func TestFlattenNode(t *testing.T) {
	roots := buildTestHierarchy()

	tests := []struct {
		name        string
		node        *TodoNode
		expectedIDs []string
	}{
		{
			name:        "leaf node",
			node:        roots[0].Children[0].Children[0], // grandchild1
			expectedIDs: []string{"grandchild1"},
		},
		{
			name:        "node with children",
			node:        roots[0].Children[0], // child1
			expectedIDs: []string{"child1", "grandchild1"},
		},
		{
			name:        "root with all descendants",
			node:        roots[0],
			expectedIDs: []string{"root1", "child1", "grandchild1", "child2"},
		},
		{
			name:        "different tree",
			node:        roots[1],
			expectedIDs: []string{"root2", "child3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenNode(tt.node)
			
			if len(result) != len(tt.expectedIDs) {
				t.Errorf("flattenNode() returned %d todos, want %d", len(result), len(tt.expectedIDs))
				return
			}

			for i, todo := range result {
				if todo.ID != tt.expectedIDs[i] {
					t.Errorf("flattenNode()[%d].ID = %s, want %s", i, todo.ID, tt.expectedIDs[i])
				}
			}
		})
	}
}

// TestHierarchyOperationsEdgeCases tests edge cases for hierarchy operations
func TestHierarchyOperationsEdgeCases(t *testing.T) {
	t.Run("operations on nil inputs", func(t *testing.T) {
		// CountHierarchyTodos with nil
		count := CountHierarchyTodos(nil)
		if count != 0 {
			t.Errorf("CountHierarchyTodos(nil) = %d, want 0", count)
		}

		// FindNodeByID with nil
		node := FindNodeByID(nil, "test")
		if node != nil {
			t.Errorf("FindNodeByID(nil, ...) expected nil")
		}

		// GetNodePath with nil
		path := GetNodePath(nil, "test")
		if path != nil {
			t.Errorf("GetNodePath(nil, ...) expected nil")
		}

		// FlattenHierarchy with nil
		todos := FlattenHierarchy(nil)
		if todos != nil {
			t.Errorf("FlattenHierarchy(nil) expected nil, got %v", todos)
		}
	})

	t.Run("deep hierarchy performance", func(t *testing.T) {
		// Create a deep hierarchy (10 levels)
		var root *TodoNode
		var current *TodoNode
		
		for i := 0; i < 10; i++ {
			node := createTestNode(string(rune('a'+i)), "Task", "")
			if root == nil {
				root = node
				current = root
			} else {
				current.Children = append(current.Children, node)
				current = node
			}
		}

		// Test operations on deep hierarchy
		count := countNodeTodos(root)
		if count != 10 {
			t.Errorf("countNodeTodos() on deep hierarchy = %d, want 10", count)
		}

		// Find deepest node
		found := findNodeInTree(root, "j")
		if found == nil {
			t.Errorf("findNodeInTree() failed to find deepest node")
		}

		// Get path to deepest node
		path := getPathToNode(root, "j", []*TodoNode{})
		if len(path) != 10 {
			t.Errorf("getPathToNode() path length = %d, want 10", len(path))
		}
	})
}

// BenchmarkHierarchyOperations benchmarks hierarchy operations
func BenchmarkHierarchyOperations(b *testing.B) {
	// Create a large hierarchy for benchmarking
	roots := make([]*TodoNode, 10)
	for i := range roots {
		root := createTestNode("root", "Root", "")
		for j := 0; j < 10; j++ {
			child := createTestNode("child", "Child", "root")
			for k := 0; k < 10; k++ {
				grandchild := createTestNode("grandchild", "Grandchild", "child")
				child.Children = append(child.Children, grandchild)
			}
			root.Children = append(root.Children, child)
		}
		roots[i] = root
	}

	b.Run("CountHierarchyTodos", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CountHierarchyTodos(roots)
		}
	})

	b.Run("FindNodeByID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindNodeByID(roots, "grandchild")
		}
	})

	b.Run("FlattenHierarchy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FlattenHierarchy(roots)
		}
	})
}