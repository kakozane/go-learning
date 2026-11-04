package main

import "fmt"

type Course struct {
	ID        string
	Name      string
	Completed bool
}

func main() {

	courses := []Course{
		{
			ID:        "1",
			Name:      "学习 Go 结构体",
			Completed: false,
		},
		{
			ID:        "2",
			Name:      "学习 Go 切片",
			Completed: false,
		},
	}

	// fmt.Println(len(courses))
	// fmt.Println(courses[0].Name)

	for i := range courses {
		courses[i].Completed = true
		fmt.Println(courses[i].Name, courses[i].Completed)

	}

}
