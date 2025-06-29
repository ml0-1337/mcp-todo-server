package core

import (
	"fmt"
	"sort"
)

// TodoNode represents a todo in a hierarchical tree structure
type TodoNode struct {
	Todo     *Todo
	Children []*TodoNode
}

// BuildTodoHierarchy builds a hierarchical tree structure from a flat list of todos
func BuildTodoHierarchy(todos []*Todo) ([]*TodoNode, []*Todo) {
	// Create a map for quick lookup
	todoMap := make(map[string]*Todo)
	nodeMap := make(map[string]*TodoNode)
	
	// Initialize nodes
	for _, todo := range todos {
		todoMap[todo.ID] = todo
		nodeMap[todo.ID] = &TodoNode{
			Todo:     todo,
			Children: []*TodoNode{},
		}
	}
	
	// Build parent-child relationships
	var roots []*TodoNode
	var orphans []*Todo
	visitedInPath := make(map[string]bool) // For circular reference detection
	
	for _, todo := range todos {
		node := nodeMap[todo.ID]
		
		if todo.ParentID == "" {
			// This is a root node
			roots = append(roots, node)
		} else {
			// Check if parent exists
			if parentNode, exists := nodeMap[todo.ParentID]; exists {
				// Check for circular reference
				if !hasCircularReference(todo.ID, todo.ParentID, todoMap, visitedInPath) {
					parentNode.Children = append(parentNode.Children, node)
				} else {
					// Treat as orphan if circular reference detected
					orphans = append(orphans, todo)
				}
			} else {
				// Parent doesn't exist - this is an orphan
				// Special handling for phase/subtask types
				if todo.Type == "phase" || todo.Type == "subtask" {
					orphans = append(orphans, todo)
				} else {
					// Non-phase/subtask todos with missing parents are treated as roots
					roots = append(roots, node)
				}
			}
		}
	}
	
	// Sort roots and children for consistent display
	sortNodes(roots)
	for _, node := range nodeMap {
		sortNodes(node.Children)
	}
	
	return roots, orphans
}

// hasCircularReference checks if adding a parent-child relationship would create a cycle
func hasCircularReference(childID, parentID string, todoMap map[string]*Todo, visited map[string]bool) bool {
	// Clear visited map for this check
	for k := range visited {
		delete(visited, k)
	}
	
	current := parentID
	for current != "" {
		if current == childID {
			return true // Found circular reference
		}
		if visited[current] {
			return false // Already visited this node, no cycle through childID
		}
		visited[current] = true
		
		if todo, exists := todoMap[current]; exists {
			current = todo.ParentID
		} else {
			break
		}
	}
	
	return false
}

// sortNodes sorts nodes by status (in_progress first), then priority (high to low), then by ID
func sortNodes(nodes []*TodoNode) {
	sort.Slice(nodes, func(i, j int) bool {
		a, b := nodes[i].Todo, nodes[j].Todo
		
		// Status priority: in_progress > blocked > completed
		statusPriority := map[string]int{
			"in_progress": 0,
			"blocked":     1,
			"completed":   2,
		}
		
		aPriority := statusPriority[a.Status]
		bPriority := statusPriority[b.Status]
		
		if aPriority != bPriority {
			return aPriority < bPriority
		}
		
		// Priority: high > medium > low
		priorityValue := map[string]int{
			"high":   0,
			"medium": 1,
			"low":    2,
		}
		
		aValue := priorityValue[a.Priority]
		bValue := priorityValue[b.Priority]
		
		if aValue != bValue {
			return aValue < bValue
		}
		
		// Finally sort by ID
		return a.ID < b.ID
	})
}

// GetHierarchyDepth returns the maximum depth of the hierarchy
func GetHierarchyDepth(roots []*TodoNode) int {
	maxDepth := 0
	for _, root := range roots {
		depth := getNodeDepth(root, 1)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

// getNodeDepth recursively calculates the depth of a node
func getNodeDepth(node *TodoNode, currentDepth int) int {
	if len(node.Children) == 0 {
		return currentDepth
	}
	
	maxChildDepth := currentDepth
	for _, child := range node.Children {
		childDepth := getNodeDepth(child, currentDepth+1)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}
	
	return maxChildDepth
}

// CountHierarchyTodos returns the total number of todos in the hierarchy
func CountHierarchyTodos(roots []*TodoNode) int {
	count := 0
	for _, root := range roots {
		count += countNodeTodos(root)
	}
	return count
}

// countNodeTodos recursively counts todos in a node and its children
func countNodeTodos(node *TodoNode) int {
	count := 1 // Count this node
	for _, child := range node.Children {
		count += countNodeTodos(child)
	}
	return count
}

// FindNodeByID finds a node by todo ID in the hierarchy
func FindNodeByID(roots []*TodoNode, id string) *TodoNode {
	for _, root := range roots {
		if node := findNodeInTree(root, id); node != nil {
			return node
		}
	}
	return nil
}

// findNodeInTree recursively searches for a node in the tree
func findNodeInTree(node *TodoNode, id string) *TodoNode {
	if node.Todo.ID == id {
		return node
	}
	
	for _, child := range node.Children {
		if found := findNodeInTree(child, id); found != nil {
			return found
		}
	}
	
	return nil
}

// GetNodePath returns the path from root to the specified node
func GetNodePath(roots []*TodoNode, id string) []*TodoNode {
	for _, root := range roots {
		if path := getPathToNode(root, id, []*TodoNode{}); path != nil {
			return path
		}
	}
	return nil
}

// getPathToNode recursively builds the path to a node
func getPathToNode(node *TodoNode, targetID string, currentPath []*TodoNode) []*TodoNode {
	currentPath = append(currentPath, node)
	
	if node.Todo.ID == targetID {
		return currentPath
	}
	
	for _, child := range node.Children {
		if path := getPathToNode(child, targetID, currentPath); path != nil {
			return path
		}
	}
	
	return nil
}

// FlattenHierarchy returns a flat list of todos from the hierarchy in display order
func FlattenHierarchy(roots []*TodoNode) []*Todo {
	var todos []*Todo
	for _, root := range roots {
		todos = append(todos, flattenNode(root)...)
	}
	return todos
}

// flattenNode recursively flattens a node and its children
func flattenNode(node *TodoNode) []*Todo {
	todos := []*Todo{node.Todo}
	for _, child := range node.Children {
		todos = append(todos, flattenNode(child)...)
	}
	return todos
}

// GetOrphanedPhases returns todos that are phases/subtasks without valid parents
func GetOrphanedPhases(todos []*Todo) []*Todo {
	// Create a set of all todo IDs for quick lookup
	todoIDs := make(map[string]bool)
	for _, todo := range todos {
		todoIDs[todo.ID] = true
	}
	
	var orphans []*Todo
	for _, todo := range todos {
		// Check if this is a phase or subtask with a parent_id that doesn't exist
		if (todo.Type == "phase" || todo.Type == "subtask") && todo.ParentID != "" {
			if !todoIDs[todo.ParentID] {
				orphans = append(orphans, todo)
			}
		}
	}
	
	return orphans
}

// HierarchyStats contains statistics about the todo hierarchy
type HierarchyStats struct {
	TotalRoots      int            `json:"total_roots"`
	TotalOrphans    int            `json:"total_orphans"`
	MaxDepth        int            `json:"max_depth"`
	TotalWithParent int            `json:"total_with_parent"`
	ByType          map[string]int `json:"by_type"`
	ByStatus        map[string]int `json:"by_status"`
}

// GetHierarchyStats calculates statistics about the todo hierarchy
func GetHierarchyStats(todos []*Todo) *HierarchyStats {
	roots, orphans := BuildTodoHierarchy(todos)
	
	stats := &HierarchyStats{
		TotalRoots:   len(roots),
		TotalOrphans: len(orphans),
		MaxDepth:     GetHierarchyDepth(roots),
		ByType:       make(map[string]int),
		ByStatus:     make(map[string]int),
	}
	
	// Count todos with parents and by type/status
	for _, todo := range todos {
		if todo.ParentID != "" {
			stats.TotalWithParent++
		}
		stats.ByType[todo.Type]++
		stats.ByStatus[todo.Status]++
	}
	
	return stats
}

// ValidateHierarchy checks for issues in the todo hierarchy
func ValidateHierarchy(todos []*Todo) []string {
	var issues []string
	
	// Check for orphaned phases/subtasks
	orphans := GetOrphanedPhases(todos)
	for _, orphan := range orphans {
		issues = append(issues, fmt.Sprintf("Orphaned %s '%s' references non-existent parent '%s'", 
			orphan.Type, orphan.ID, orphan.ParentID))
	}
	
	// Check for circular references
	visitedInPath := make(map[string]bool)
	todoMap := make(map[string]*Todo)
	for _, todo := range todos {
		todoMap[todo.ID] = todo
	}
	
	for _, todo := range todos {
		if todo.ParentID != "" {
			if hasCircularReference(todo.ID, todo.ParentID, todoMap, visitedInPath) {
				issues = append(issues, fmt.Sprintf("Circular reference detected: '%s' -> '%s'", 
					todo.ID, todo.ParentID))
			}
		}
	}
	
	// Check for phase/subtask without parent_id
	for _, todo := range todos {
		if (todo.Type == "phase" || todo.Type == "subtask") && todo.ParentID == "" {
			issues = append(issues, fmt.Sprintf("%s '%s' should have a parent_id", 
				todo.Type, todo.ID))
		}
	}
	
	return issues
}