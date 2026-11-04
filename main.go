package main

import (
	"errors"
	"fmt"
	"strings"
)

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

	// for i := range courses {
	// 	courses[i].Completed = true
	// 	fmt.Println(courses[i].Name, courses[i].Completed)

	// }

	courses = append(courses, Course{
		ID:        "3",
		Name:      "学习 Go 函数",
		Completed: false,
	})

	// for i := range courses {
	// 	fmt.Println(courses[i])
	// }

	courses = addCourse(courses, Course{
		ID:        "4",
		Name:      "学习 Go 接口",
		Completed: false,
	})
	// fmt.Print(courses)
	// course, found := findCourseByID(courses, "66")
	// if found {
	// 	fmt.Println("找到课程：", course.Name)
	// } else {
	// 	fmt.Println("课程不存在")
	// }

	err, i := updateNameByID(courses, "1", "")
	if err == nil {
		fmt.Println("修改成功：", courses[i].Name)
	} else {
		fmt.Println("修改失败:", err)
	}

	// remaining, deleted := deleteCourseByID(courses, "2")
	// if deleted {
	// 	fmt.Println("课程已删除")
	// } else {
	// 	fmt.Println("课程不存在")
	// }
	// fmt.Println("当前课程为：", remaining)
}

func addCourse(courses []Course, course Course) []Course {
	return append(courses, course)
}

// findCourseByID 根据课程 ID 查找课程
func findCourseByID(courses []Course, id string) (Course, bool) {
	for _, course := range courses {
		if course.ID == id {
			// 结束
			return course, true
		}
	}
	// 没找到的情况
	return Course{}, false
}

// updateNameByID 根据课程 ID 更新课程名称
func updateNameByID(courses []Course, id string, newName string) (error, int) {

	name := strings.TrimSpace(newName)

	if name == "" {
		return errors.New("课程名称不能为空"), -1
	}

	for i, course := range courses {
		if course.ID == id {
			courses[i].Name = name
			return nil, i
		}
	}
	return errors.New("课程不存在"), -1
}

// deleteCourseByID 根据课程 ID 删除课程
func deleteCourseByID(courses []Course, id string) ([]Course, bool) {

	for i, course := range courses {
		if course.ID == id {
			newCourses := append(courses[:i], courses[i+1:]...)
			return newCourses, true
		}
	}
	return courses, false
}
