package main

import (
	"fmt"

	xmldot "github.com/netascode/xmldot"
)

// Sample XML: Employee database
const employeesXML = `<company>
	<employees>
		<employee id="1001" status="active">
			<name>Alice</name>
			<age>28</age>
			<department>Engineering</department>
			<salary>85000</salary>
		</employee>
		<employee id="1002" status="active">
			<name>Bob</name>
			<age>35</age>
			<department>Sales</department>
			<salary>75000</salary>
		</employee>
		<employee id="1003" status="inactive">
			<name>Carol</name>
			<age>42</age>
			<department>Engineering</department>
			<salary>95000</salary>
		</employee>
		<employee id="1004" status="active">
			<name>David</name>
			<age>31</age>
			<department>Marketing</department>
			<salary>70000</salary>
		</employee>
		<employee id="1005" status="active">
			<name>Eve</name>
			<age>29</age>
			<department>Engineering</department>
			<salary>88000</salary>
		</employee>
	</employees>
</company>`

func main() {
	fmt.Println("Filter Operations Example")
	fmt.Println("=========================\n")

	// Example 1: Numeric filter (greater than) - all matches
	fmt.Println("Example 1: Employees older than 30")
	result := xmldot.Get(employeesXML, "company.employees.employee.#(age>30)#.name")
	for _, name := range result.Array() {
		fmt.Printf("  - %s\n", name.String())
	}
	fmt.Println()

	// Example 2: String filter (equality) - all matches
	fmt.Println("Example 2: Employees in Engineering")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(department==Engineering)#.name")
	for _, name := range result.Array() {
		fmt.Printf("  - %s\n", name.String())
	}
	fmt.Println()

	// Example 3: Attribute filter - all matches
	fmt.Println("Example 3: Active employees")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(@status==active)#.name")
	for _, name := range result.Array() {
		fmt.Printf("  - %s\n", name.String())
	}
	fmt.Println()

	// Example 4: Comparison operators - all matches
	fmt.Println("Example 4: Salary ranges")
	fmt.Println("High earners (>= $85,000):")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(salary>=85000)#.name")
	for _, name := range result.Array() {
		fmt.Printf("  - %s\n", name.String())
	}
	fmt.Println()

	fmt.Println("Low to mid earners (< $80,000):")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(salary<80000)#.name")
	for _, name := range result.Array() {
		fmt.Printf("  - %s\n", name.String())
	}
	fmt.Println()

	// Example 5: Not equal operator - all matches
	fmt.Println("Example 5: Non-Sales employees")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(department!=Sales)#.name")
	for _, name := range result.Array() {
		fmt.Printf("  - %s\n", name.String())
	}
	fmt.Println()

	// Example 6: Manual filtering (chained filters not supported)
	fmt.Println("Example 6: Active Engineering employees over 30")
	// Note: Chained filters like #(...)#(...)#(...) don't work - use manual filtering
	activeEmps := xmldot.Get(employeesXML, "company.employees.employee.#(@status==active)#")
	activeEmps.ForEach(func(index int, emp xmldot.Result) bool {
		dept := xmldot.Get(emp.Raw, "department")
		age := xmldot.Get(emp.Raw, "age")
		if dept.String() == "Engineering" && age.Int() > 30 {
			name := xmldot.Get(emp.Raw, "name")
			fmt.Printf("  - %s\n", name.String())
		}
		return true
	})
	fmt.Println()

	// Example 7: Filter with iteration - all matches
	fmt.Println("Example 7: All active employees' details")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(@status==active)#")
	result.ForEach(func(index int, value xmldot.Result) bool {
		name := xmldot.Get(value.Raw, "name")
		dept := xmldot.Get(value.Raw, "department")
		fmt.Printf("  - %s (%s)\n", name.String(), dept.String())
		return true
	})
	fmt.Println()

	// Example 8: No matches
	fmt.Println("Example 8: Employees under 25 (no matches)")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(age<25)#.name")
	if !result.Exists() || len(result.Array()) == 0 {
		fmt.Println("  No employees found")
	}
	fmt.Println()

	// Example 9: Complex query with modifiers
	fmt.Println("Example 9: Top 2 earners (sorted)")
	result = xmldot.Get(employeesXML, "company.employees.employee.salary|@sort|@reverse|@first")
	fmt.Printf("Highest salary: $%s\n", result.String())
	fmt.Println()

	// Example 10: Count filtered results
	fmt.Println("Example 10: Count active employees")
	result = xmldot.Get(employeesXML, "company.employees.employee.#(@status==active)#")
	fmt.Printf("Active employees: %d\n", len(result.Array()))
}
