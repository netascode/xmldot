package main

import (
	"fmt"
	"log"

	xmldot "github.com/netascode/xmldot"
)

// Sample XML: User profile
const profileXML = `<user>
	<name>John Doe</name>
	<email>john@example.com</email>
	<settings>
		<theme>light</theme>
		<notifications enabled="true"/>
	</settings>
</user>`

func main() {
	fmt.Println("Basic Set Operations Example")
	fmt.Println("============================\n")

	// Example 1: Update element value
	fmt.Println("Example 1: Update element value")
	xml, err := xmldot.Set(profileXML, "user.name", "Jane Smith")
	if err != nil {
		log.Fatal(err)
	}
	name := xmldot.Get(xml, "user.name")
	fmt.Printf("Updated name: %s\n\n", name.String())

	// Example 2: Set attribute
	fmt.Println("Example 2: Set attribute")
	xml, err = xmldot.Set(xml, "user.settings.notifications.@enabled", "false")
	if err != nil {
		log.Fatal(err)
	}
	enabled := xmldot.Get(xml, "user.settings.notifications.@enabled")
	fmt.Printf("Notifications enabled: %s\n\n", enabled.String())

	// Example 3: Create new element
	fmt.Println("Example 3: Create new element")
	xml, err = xmldot.Set(xml, "user.phone", "555-1234")
	if err != nil {
		log.Fatal(err)
	}
	phone := xmldot.Get(xml, "user.phone")
	fmt.Printf("Phone: %s\n\n", phone.String())

	// Example 4: Create nested path
	fmt.Println("Example 4: Create nested path")
	xml, err = xmldot.Set(xml, "user.settings.privacy.level", "high")
	if err != nil {
		log.Fatal(err)
	}
	privacy := xmldot.Get(xml, "user.settings.privacy.level")
	fmt.Printf("Privacy level: %s\n\n", privacy.String())

	// Example 5: Delete element
	fmt.Println("Example 5: Delete element")
	xml, err = xmldot.Delete(xml, "user.phone")
	if err != nil {
		log.Fatal(err)
	}
	phone = xmldot.Get(xml, "user.phone")
	fmt.Printf("Phone exists: %v\n\n", phone.Exists())

	// Example 6: Delete attribute
	fmt.Println("Example 6: Delete attribute")
	xml, err = xmldot.Delete(xml, "user.settings.notifications.@enabled")
	if err != nil {
		log.Fatal(err)
	}
	enabled = xmldot.Get(xml, "user.settings.notifications.@enabled")
	fmt.Printf("Enabled attribute exists: %v\n\n", enabled.Exists())

	// Example 7: Set with different types
	fmt.Println("Example 7: Set with different types")
	xml, err = xmldot.Set(profileXML, "user.age", 30)
	if err != nil {
		log.Fatal(err)
	}
	xml, err = xmldot.Set(xml, "user.premium", true)
	if err != nil {
		log.Fatal(err)
	}
	age := xmldot.Get(xml, "user.age")
	premium := xmldot.Get(xml, "user.premium")
	fmt.Printf("Age: %d\n", age.Int())
	fmt.Printf("Premium: %v\n\n", premium.Bool())

	// Example 8: Verify final state
	fmt.Println("Example 8: Final XML state")
	finalXML := xmldot.Get(xml, "user|@pretty")
	fmt.Println(finalXML.String())
}
