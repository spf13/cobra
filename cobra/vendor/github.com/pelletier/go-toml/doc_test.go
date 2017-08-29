// code examples for godoc

package toml

import (
	"fmt"
)

func Example_tree() {
	config, err := LoadFile("config.toml")

	if err != nil {
		fmt.Println("Error ", err.Error())
	} else {
		// retrieve data directly
		user := config.Get("postgres.user").(string)
		password := config.Get("postgres.password").(string)

		// or using an intermediate object
		configTree := config.Get("postgres").(*Tree)
		user = configTree.Get("user").(string)
		password = configTree.Get("password").(string)
		fmt.Println("User is", user, " and password is", password)

		// show where elements are in the file
		fmt.Printf("User position: %v\n", configTree.GetPosition("user"))
		fmt.Printf("Password position: %v\n", configTree.GetPosition("password"))
	}
}

func Example_unmarshal() {
	type Employer struct {
		Name  string
		Phone string
	}
	type Person struct {
		Name     string
		Age      int64
		Employer Employer
	}

	document := []byte(`
	name = "John"
	age = 30
	[employer]
		name = "Company Inc."
		phone = "+1 234 567 89012"
	`)

	person := Person{}
	Unmarshal(document, &person)
	fmt.Println(person.Name, "is", person.Age, "and works at", person.Employer.Name)
}
