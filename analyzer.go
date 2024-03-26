package main

import "fmt"

func (c *Checker) topologicalSort(stmts []*VarDefStmt) []*VarDefStmt {
	m := c.getDependencyMatrix(stmts)

	visited := make(map[string]bool)
	stack := &Stack[string]{}

	for k, _ := range m {
		if !visited[k] {
			c.topologicalSortHelper(k, m, visited, stack)
		}
	}

	sortedList := make([]*VarDefStmt, 0)
	sortedVariables := stack.Reverse()
	i := 0
	for !sortedVariables.Empty() {
		name := sortedVariables.Pop()
		if c.GetVariable(name) != nil {
			stmt := c.GetVariable(name).(*VarDef).DefNode
			stmt.Order = i
			sortedList = append(sortedList, stmt)
			i++
		}
	}
	return sortedList
}

func (c *Checker) topologicalSortHelper(key string, m map[string][]string, visited map[string]bool, stack *Stack[string]) {
	visited[key] = true
	for _, v := range m[key] {
		if !visited[v] {
			c.topologicalSortHelper(v, m, visited, stack)
		}
	}

	stack.Push(key)
}

func (c *Checker) getDependencyMatrix(stmts []*VarDefStmt) map[string][]string {
	var m = make(map[string][]string)
	for _, s := range stmts {
		if _, ok := m[s.Name.Name]; !ok {
			m[s.Name.Name] = make([]string, 0)
		}

		if s.Init == nil {
			continue
		}
		m[s.Name.Name] = c.collectDependenciesOfInit(s.Init)
	}

	return m
}

func (c *Checker) checkInitializationCycle(stmts []*VarDefStmt) bool {
	/*
		var a = b
		var b = c
		var c = a
		a -> b -> c ->
	*/
	var m = c.getDependencyMatrix(stmts)
	var visited = make(map[string]bool)
	var currentPath = make(map[string]bool)

	for k, _ := range m {
		visited[k] = false
		currentPath[k] = false
	}

	for k, _ := range m {
		if !visited[k] {
			if ok, nameList := c.detectCycle(m, k, visited, currentPath); ok {
				c.addCycleErrors(nameList)
				return true
			}
		}
	}

	return false

}

func (c *Checker) addCycleErrors(nameList []string) {
	for i, name := range nameList {
		variable := c.GetVariable(name)
		if variable != nil && variable.IsVarDef() {
			start, end := variable.Pos()
			if i == 0 {
				c.errorf(start, end, "initialization cycle detected")
			}

			c.errorf(start, end, fmt.Sprintf("%s depends on  ", name))

		}
	}
	variable := c.GetVariable(nameList[0])
	if variable != nil && variable.IsVarDef() {
		start, end := variable.Pos()
		c.errorf(start, end, nameList[0])
	}
}

func (c *Checker) detectCycle(m map[string][]string, key string, visited map[string]bool, currentPath map[string]bool) (bool, []string) {
	visited[key] = true
	currentPath[key] = true
	defer func() { currentPath[key] = false }()
	var orders = make([]string, 0)
	orders = append(orders, key)

	for _, v := range m[key] {
		if !visited[v] {
			ok, orders2 := c.detectCycle(m, v, visited, currentPath)
			if ok {
				orders = append(orders, orders2...)
				return true, orders
			}
		} else if currentPath[v] {
			return true, orders
		}
	}

	return false, make([]string, 0)
}

func (c *Checker) collectDependenciesOfInit(init Expr) []string {
	resultList := make([]string, 0)
	switch init := init.(type) {
	case *IdentExpr:
		resultList = append(resultList, init.Name)
	case *BinaryExpr:
		leftList := c.collectDependenciesOfInit(init.Left)
		rightList := c.collectDependenciesOfInit(init.Right)
		resultList = append(resultList, leftList...)
		resultList = append(resultList, rightList...)
	case *UnaryExpr:
		resultList = append(resultList, c.collectDependenciesOfInit(init.Right)...)
	case *SelectorExpr:
		panic("unsupported feature yet")
	default:
		// continue
	}

	return resultList
}
