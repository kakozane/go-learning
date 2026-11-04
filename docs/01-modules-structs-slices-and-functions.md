# 第一节 Go 入门：模块、结构体、切片与函数

本节围绕课程管理小练习，学习程序入口、Course 结构体、切片遍历与追加，以及函数参数的值传递。

## 1 初始化模块并运行程序

在项目根目录执行：

```bash
go mod init github.com/kakozane/go-learning
```

该命令生成 go.mod，声明模块路径和 Go 版本。仓库已经初始化过，不需要重复执行。模块路径使用 GitHub 仓库地址形式，不代表自动上传代码或安装依赖。

最小程序 main.go：

```go
package main

import "fmt"

func main() {
	fmt.Println("开始学习 Go")
}
```

- package main：可执行程序的入口包。
- import "fmt"：导入标准库，无需额外安装。
- func main()：程序入口函数。
- fmt.Println：打印并换行；fmt.Print 不自动换行。

运行当前目录的包：

```bash
go run .
```

## 2 定义 Course 结构体

```go
type Course struct {
	ID        string
	Name      string
	Completed bool
}
```

Course 是自定义类型，每个值表示一门课程。字段名在前、类型在后，不写冒号。

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| ID | string | 课程 ID |
| Name | string | 课程名称 |
| Completed | bool | 是否完成 |

字段首字母大写表示可导出，其他包可以访问。Go 中通常将标识符中的 ID 保持全大写。

与 TypeScript 的对象类型相比，两者都能描述字段，但 Go 结构体是值类型，不能照搬 JavaScript 对象引用的理解。

## 3 创建、读取和修改结构体

```go
course := Course{
	ID:        "1",
	Name:      "学习 Go 结构体",
	Completed: false,
}

fmt.Println(course.Name)
course.Completed = true
fmt.Println(course.Completed)
```

`:=` 是短变量声明，编译器根据右侧推断变量类型。

结构体字面量使用 `字段名: 值`。多行写法每个元素后保留逗号，包括最后一项。

省略的字段取对应类型的零值：string 是空字符串，bool 是 false。因此下面的 Completed 也是 false：

```go
course := Course{ID: "1", Name: "学习 Go"}
```

普通 Go 结构体字段可以直接修改，没有 React state 的不可变更新约束。

## 4 用切片保存多门课程

```go
courses := []Course{
	{ID: "1", Name: "学习 Go 结构体"},
	{ID: "2", Name: "学习 Go 切片"},
}

fmt.Println(len(courses))
fmt.Println(courses[0].Name)
```

Course 表示单个课程类型，[]Course 表示元素类型为 Course 的切片。

外层已说明元素类型，内层结构体字面量可以省略重复的 Course。

len 获取长度，下标从 0 开始。访问下标必须小于长度；空切片不能直接访问 courses[0]。

切片和数组不是同一种类型：[]Course 是切片，[2]Course 是长度为 2 的数组。本节先使用切片。

## 5 range 遍历与下划线

```go
for i, course := range courses {
	fmt.Println(i, course.Name, course.Completed)
}
```

第一个值是下标，第二个值是当前元素的值。

| 写法 | 用途 |
| --- | --- |
| `for i := range courses` | 只需要下标 |
| `for _, course := range courses` | 只使用元素值 |
| `for i, course := range courses` | 同时使用下标与元素值 |

下划线是空白标识符，表示忽略值。只需要第二个值时，前面的下划线不能省略，否则拿到的是下标。

```go
for i, _ := range courses {
	fmt.Println(i)
}
```

上面合法，但工具会给出 simplifyrange 简化提示，因为末尾的下划线可以省略：

```go
for i := range courses {
	fmt.Println(i)
}
```

这是简化建议，不是逻辑错误。

## 6 为什么 range 中的 course 是副本

Go 的赋值复制值。对于 []Course，range 将当前结构体元素的值复制给循环变量：

```go
for _, course := range courses {
	course.Completed = true
}
```

可以把每轮理解为：

```go
course := courses[i]
course.Completed = true
```

只改变了局部结构体副本中的 bool 字段，没有改变切片元素。

单独看结构体赋值也一样：

```go
original := Course{Name: "学习 Go"}
copied := original
copied.Completed = true

fmt.Println(original.Completed) // false
fmt.Println(copied.Completed)   // true
```

修改原切片元素时，使用下标：

```go
for i := range courses {
	courses[i].Completed = true
	fmt.Println(courses[i].Name, courses[i].Completed)
}
```

与 JavaScript 的对象赋值不同，Go 这里复制的是结构体值。结构体复制不是通用的深拷贝；如果以后字段包含指针、切片或 map，还需要理解这些字段的共享关系。本节 Course 的字段足以直接用上述规则解释。

## 7 append 追加元素

```go
courses = append(courses, Course{
	ID:   "3",
	Name: "学习 Go 函数",
})
```

append 返回追加后的切片，需要接住返回值。

它不保证创建独立的新底层数组：容量够时可以复用原数组，容量不足时会分配新数组并复制原有元素。因此不要把它等同于 React 中的数组展开更新。

## 8 将添加操作提取为函数

```go
func addCourse(courses []Course, course Course) []Course {
	return append(courses, course)
}
```

| 部分 | 含义 |
| --- | --- |
| courses []Course | 接收已有课程切片 |
| course Course | 接收新课程 |
| 参数列表后的 []Course | 返回值类型 |
| return | 将追加后的切片返回给调用方 |

在 main 中调用：

```go
courses = addCourse(courses, Course{
	ID:   "4",
	Name: "学习 Go 接口",
})
```

名称中提到接口只是课程数据内容，本节还没有学习 Go interface。

## 9 为什么追加后还要返回并接住

函数参数按值传递。传入切片时，复制的是切片描述信息，可以先理解为：

```text
底层数组的位置 + 长度 len + 容量 cap
```

函数内外的切片最初指向同一个底层数组，但各自拥有自己的长度和容量信息。

修改已有元素：

```go
func completeFirst(courses []Course) {
	if len(courses) > 0 {
		courses[0].Completed = true
	}
}
```

这里修改共享底层数组中的元素，调用方能够看到。

append 则需要增加长度，容量不足时还会改变底层数组的位置。调用方的切片变量不会自动得到这些新信息，所以应当：

```go
courses = addCourse(courses, newCourse)
```

如果只调用而不接收：

```go
addCourse(courses, newCourse)
```

调用方 courses 的长度不变。即使容量够、没有重新分配数组，调用方也不会自动增加长度。

注意两个不同层面的复制：

- 函数接收切片：复制切片描述信息，可能共享底层数组。
- range 取出 Course：复制一个结构体元素。

## 10 函数定义放在哪里

普通具名函数不能直接声明在另一个函数内部：

```go
func main() {
	// 此处不能直接声明 func addCourse(...) ...
}
```

应与 main 并列：

```go
func main() {
	// 在这里调用 addCourse
}

func addCourse(courses []Course, course Course) []Course {
	return append(courses, course)
}
```

位置由大括号决定，不由缩进决定。addCourse 可以定义在 main 后面，不需要提前声明。

之前的 `expected '(', found addCourse` 报错，实际原因就是函数定义放在 main 内部。Go 支持函数值，但本节先使用包级函数。

## 11 当前练习的完整简化版

```go
package main

import "fmt"

type Course struct {
	ID        string
	Name      string
	Completed bool
}

func main() {
	courses := []Course{
		{ID: "1", Name: "学习 Go 结构体"},
		{ID: "2", Name: "学习 Go 切片"},
	}

	courses = append(courses, Course{
		ID:   "3",
		Name: "学习 Go 函数",
	})

	courses = addCourse(courses, Course{
		ID:   "4",
		Name: "学习 Go 接口",
	})

	for _, course := range courses {
		fmt.Println(course.ID, course.Name, course.Completed)
	}
}

func addCourse(courses []Course, course Course) []Course {
	return append(courses, course)
}
```

示例为便于阅读按行打印；实际 main.go 使用 fmt.Print(courses) 输出整个切片，两者都能观察添加结果。

要观察添加第 4 门课程后的结果，打印应放在 addCourse 调用之后。

## 12 本节练习与验证

1. 添加第 5 门课程，并打印最终数量。
2. 用下标把某门课程标记为完成。
3. 对比修改 range 副本和修改 courses[i] 的输出。
4. 临时丢弃 addCourse 返回值，观察外部长度不增加，再恢复正确写法。
5. 解释 append 追加与修改已有元素的不同。

检查命令：

```bash
gofmt -l main.go
go run .
go vet ./...
go test ./...
```

本次实际检查：main.go 格式通过，程序成功输出四门课程，addCourse 已移至 main 外；go vet 和 go test 执行成功。目前没有测试文件，不能据此声称业务逻辑已有自动化测试覆盖。

下一步：按 ID 查找课程，学习 Go 的多返回值与 found 标志。
