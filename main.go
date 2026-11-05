package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 告诉 JSON 编码器使用什么字段名
type Course struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}

type Task struct {
	Title string
	Done  bool
}

type ResData struct {
	Code    int      `json:"code"`
	Data    []Course `json:"data"`
	Message string   `json:"message"`
}

type Completer interface {
	Complete()
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
		{
			ID:        "3",
			Name:      "学习 Go 函数",
			Completed: false,
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /courses", func(w http.ResponseWriter, r *http.Request) {

		resData := ResData{
			Code:    200,
			Data:    courses,
			Message: "请求成功！",
		}

		data, err := json.Marshal(resData)

		if err != nil {
			http.Error(w, "课程数据编码失败", http.StatusInternalServerError)
			return
		}

		fmt.Println(string(data))

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(data)

		if err != nil {
			log.Println("响应写入失败：", err)
		}
	})

	fmt.Println("准备启动服务：http://localhost:8080")
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		log.Fatal("服务启动失败：", err)
	}
}

// func addCourse(courses []Course, course Course) []Course {
// 	return append(courses, course)
// }

// findCourseByID 根据课程 ID 查找课程
// func findCourseByID(courses []Course, id string) (Course, bool) {
// 	for _, course := range courses {
// 		if course.ID == id {
// 			// 结束
// 			return course, true
// 		}
// 	}
// 	// 没找到的情况
// 	return Course{}, false
// }

// func findCoursePointerByID(courses []Course, id string) *Course {
// 	for i := range courses {
// 		if courses[i].ID == id {
// 			return &courses[i]
// 		}
// 	}
// 	return nil
// }

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

func completeCourse(course *Course) {
	course.Completed = true
}

func (c *Course) Complete() {
	c.Completed = true
}

func (t *Task) Complete() {
	t.Done = true
}

func completeItem(item Completer) {
	item.Complete()
}

func completeAll(items []Completer) {
	for _, item := range items {
		item.Complete()
	}
}

// func main() {

// 	courses := []Course{
// 		{
// 			ID:        "1",
// 			Name:      "学习 Go 结构体",
// 			Completed: false,
// 		},
// 		{
// 			ID:        "2",
// 			Name:      "学习 Go 切片",
// 			Completed: false,
// 		},
// 	}

// 	courses = append(courses, Course{
// 		ID:        "3",
// 		Name:      "学习 Go 函数",
// 		Completed: false,
// 	})

// fmt.Println(len(courses))
// fmt.Println(courses[0].Name)

// for i := range courses {
// 	courses[i].Completed = true
// 	fmt.Println(courses[i].Name, courses[i].Completed)

// }

// course := Course{
// 	ID:        "4",
// 	Name:      "学习 Go 指针",
// 	Completed: false,
// }

// course.Complete()

// fmt.Println(course)
// for i := range courses {
// 	fmt.Println(courses[i])
// }

// fmt.Print(courses)
// course, found := findCourseByID(courses, "1")
// if found {
// 	fmt.Println("找到课程：", course.Name)
// 	course.Name = "修改课程名称"
// } else {
// 	fmt.Println("课程不存在")
// }

// err, i := updateNameByID(courses, "1", "")
// if err == nil {
// 	fmt.Println("修改成功：", courses[i].Name)
// } else {
// 	fmt.Println("修改失败:", err)
// }

// remaining, deleted := deleteCourseByID(courses, "2")
// if deleted {
// 	fmt.Println("课程已删除")
// } else {
// 	fmt.Println("课程不存在")
// }
// fmt.Println("当前课程为：", remaining)

// 指针练习
// course := Course{
// 	ID:   "5",
// 	Name: "学习 Go 指针",
// }
// completeCourse(&course)

// fmt.Println(course.Completed)

// coursePtr := findCoursePointerByID(courses, "1")
// if coursePtr == nil {
// 	fmt.Println("课程不存在")
// } else {
// 	coursePtr.Name = "通过指针修改的课程"
// 	fmt.Println("修改后的课程名称为：", coursePtr.Name)
// }
// fmt.Println("原始切片为：", courses)

// demoCourse := Course{
// 	ID:   "100",
// 	Name: "学习 Go 接口",
// }

// demoTask := Task{
// 	Title: "整理学习笔记",
// }

// fmt.Println("完成前：", demoCourse.Completed, demoTask.Done)

// completeAll([]Completer{&demoCourse, &demoTask})

// fmt.Println("完成后：", demoCourse.Completed, demoTask.Done)

// 	listData, err := json.Marshal(courses)
// 	if err != nil {
// 		fmt.Println("JSON 编码失败", err)
// 		return
// 	}
// 	fmt.Println("JSON 编码后的数据：", string(listData))

// 	var decodedCourses []Course

// 	err = json.Unmarshal(listData, &decodedCourses)
// 	if err != nil {
// 		fmt.Println("JSON解码失败", err)
// 		return
// 	}

// 	for _, course := range decodedCourses {
// 		fmt.Println("json解码数据", course.ID, course.Name, course.Completed)
// 	}

// }
