# 第二节 课程查找、修改与删除

本节完成按 ID 查找、修改名称与删除课程，学习多返回值、切片截取、参数展开及底层数组共享。

上一篇：[第一节 模块、结构体、切片与函数](./01-modules-structs-slices-and-functions.md)。

## 1 查找课程与多返回值

```go
func findCourseByID(courses []Course, id string) (Course, bool) {
	for _, course := range courses {
		if course.ID == id {
			return course, true
		}
	}
	return Course{}, false
}
```

返回类型 `(Course, bool)` 表示返回课程值和是否找到。找到后 return 立即结束整个函数，不需要 break。

Course{} 是字段均为零值的结构体，ID 和 Name 为空字符串，Completed 为 false。需要通过第二个返回值判断是否找到，不应只看课程内容。

```go
course, found := findCourseByID(courses, "2")
if found {
	fmt.Println("找到课程：", course.Name)
} else {
	fmt.Println("课程不存在")
}
```

接收变量名可以自己选择，但顺序必须对应。返回的 Course 仍是结构体副本，直接修改它不会改变原切片元素。

## 2 根据 ID 修改课程名称

```go
func updateNameByID(courses []Course, id string, newName string) (bool, int) {
	for i, course := range courses {
		if course.ID == id {
			courses[i].Name = newName
			return true, i
		}
	}
	return false, -1
}
```

用 course 副本读取 ID，用 courses[i] 修改原元素，这两者并不冲突。

```go
updated, index := updateNameByID(courses, "2", "深入学习 Go 切片")
if updated {
	fmt.Println("修改成功：", courses[index].Name)
} else {
	fmt.Println("课程不存在")
}
```

合法下标从 0 开始，-1 用于表示没找到。必须先判断 updated，不能直接用 -1 访问切片，否则会 panic。

如果调用方只需要成功与否，返回 bool 即可；当前练习希望得到修改位置，所以返回 bool 和 int。

修改已有元素不改变切片长度，不需要像 append 一样返回更新后的切片描述信息。本节未添加空名称校验，后续学习错误处理时再扩展。

## 3 切片截取：左闭右开

假设原来有四门课程：

```text
下标： 0  1  2  3
课程：[A, B, C, D]
```

要删除下标 i = 1 的 B：

```go
courses[:i]   // [A]：从开头到 i 之前，不包含 i
courses[i+1:] // [C, D]：从 i+1 到末尾
```

省略左边表示从 0 开始，省略右边表示到当前长度。截取允许边界等于长度，因此删除最后一个元素时，courses[i+1:] 是空切片。

## 4 删除函数

```go
func deleteCourseByID(courses []Course, id string) ([]Course, bool) {
	for i, course := range courses {
		if course.ID == id {
			return append(courses[:i], courses[i+1:]...), true
		}
	}
	return courses, false
}
```

找到时拼接前后两部分，返回缩短后的切片和 true。找不到时返回原切片和 false。若数据里有重复 ID，这个函数只删除第一个匹配项；当前练习按 ID 唯一的数据使用。

## 5 为什么后半部分要写省略号

append 的形式是：

```go
append(目标切片, 元素1, 元素2)
```

第一个参数本来就要求是切片，因此 courses[:i] 不展开。后面的参数要求是 Course 元素，而 courses[i+1:] 是 []Course，要用 ... 把它展开成多个元素。

```go
before := courses[:i]
after := courses[i+1:]
result := append(before, after...)
```

... 只作用于紧挨着它前面的 after，不是对整个 append 操作生效。

用整数举例：

```go
before := []int{1, 2}
after := []int{4, 5}
result := append(before, after...)
fmt.Println(result) // [1 2 4 5]
```

不使用展开时，可以逐个追加或循环：

```go
result := before
for _, value := range after {
	result = append(result, value)
}
```

直接 append(before, after) 会类型不匹配：它需要 int 元素，却收到 []int。

## 6 本次遇到的截取错误

错误写法：

```go
append(courses[:i], courses[:i+1]...)
```

后半部分写成了从开头截取，而不是从删除位置之后截取。对于 ID 为 1、2、3、4 的课程，删除 ID 2 时：

```text
错误：[1] + [1, 2] = [1, 1, 2]
正确：[1] + [3, 4] = [1, 3, 4]
```

正确的是 `courses[i+1:]`，注意冒号在 i+1 后面。

返回 true 只表示函数走到了找到分支，并不能证明返回的数据正确，所以需要检查结果内容。

## 7 会修改原来的底层数组吗

会。这个删除写法的目标切片保留了原容量，足以容纳删除后的元素，因此 append 复用底层数组，将后面的元素往前覆盖。

```go
numbers := []int{1, 2, 3, 4}
result := append(numbers[:1], numbers[2:]...)

fmt.Println(result)  // [1 3 4]
fmt.Println(numbers) // [1 3 4 4]
```

底层数组变了，但 numbers 的长度还是 4；返回的 result 长度为 3。末尾残留的值不属于 result 的有效范围。

普通 append 是否分配新数组取决于容量，不能说它总是修改原数组；但当前删除写法容量足够，会复用。

## 8 调用方如何使用结果

```go
remaining, deleted := deleteCourseByID(courses, "2")
courses = remaining

if deleted {
	fmt.Println("课程已删除")
} else {
	fmt.Println("课程不存在")
}
fmt.Println("当前课程为：", courses)
```

如果只打印 remaining 后程序就结束，不接回 courses 也可以，这正是当前 main.go 的写法。

如果后续继续使用 courses 作为课程列表，则必须接回结果，更新其长度。不要同时把旧长度的 courses 当作删除后的列表使用。

## 9 当前程序输出与验证

当前代码查找 ID 66，修改 ID 666，这两个值都不存在，因此前两条失败提示正常。删除 ID 2 后，剩余 ID 应为 1、3、4。

main_test.go 增加了小范围回归测试：

- 删除中间、第一、最后、唯一元素。
- 删除不存在的 ID、从空切片删除。
- 查找成功与失败、返回结构体为副本。
- 修改成功与失败，以及原切片元素是否正确变化。

运行：

```bash
go run .
go vet ./...
go test ./...
```

测试用于防止本次切片方向错误再次出现；测试语法之后再系统学习，不需要现在逐行掌握。

下一步可以学习 error：为修改名称增加空值校验，并区分课程不存在与名称不合法。
