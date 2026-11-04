# 第四节 指针、方法与接收者

上一篇：[第三节 error 错误处理](./03-error-handling.md)。

本节理解结构体副本与指针的区别，并给 Course 定义完成课程的方法。

## 1 Go 参数按值传递

```go
func completeByValue(course Course) {
	course.Completed = true
}
```

调用 completeByValue(myCourse) 会把结构体值复制给参数，修改参数不会改变原结构体。

```go
myCourse := Course{Name: "学习 Go 指针"}
completeByValue(myCourse)
fmt.Println(myCourse.Completed) // false
```

这与 range 的元素变量拿到 Course 副本是同样的值复制规则。

## 2 地址与指针

```go
pointer := &myCourse
```

myCourse 的类型是 Course，pointer 的类型是 *Course，保存的是指向原课程的地址。

| 写法 | 含义 |
| --- | --- |
| `*Course` | 指向 Course 的指针类型 |
| `&myCourse` | 取得变量地址 |
| `*pointer` | 通过指针访问其指向的值，称为解引用 |

指针不是另一份课程数据。即使指针值被复制，多个指针仍可以指向同一个课程。

## 3 通过指针修改原值

```go
func completeCourse(course *Course) {
	course.Completed = true
}
```

```go
myCourse := Course{Name: "学习 Go 指针"}
completeCourse(&myCourse)
fmt.Println(myCourse.Completed) // true
```

Go 仍然复制参数，只是这次复制的是指针值，不是结构体值。函数通过这个地址找到原课程并修改字段。

结构体指针访问字段可以简写：

```go
(*pointer).Completed = true
pointer.Completed = true
```

两句效果相同。指针为 nil 时不能这样访问字段，否则会 panic。

## 4 哪些副本仍会影响原数据

| 操作 | 复制什么 | 修改行为 |
| --- | --- | --- |
| range 取出 Course | 结构体元素值 | 修改副本字段不影响原元素 |
| 参数 Course | 结构体值 | 修改参数字段不影响原值 |
| 参数 *Course | 指针值 | 通过地址修改字段会影响原值 |
| 参数 []Course | 切片描述信息 | 修改共享底层数组中的元素会影响原数据 |

因此“是副本”不等于“底层数据完全独立”。range 取元素值也不等于每次循环复制整个切片。

结构体值复制不是通用深拷贝；含有指针、切片等字段时还要考虑字段自身的共享关系。

## 5 指针查找函数

```go
func findCoursePointerByID(courses []Course, id string) *Course {
	for i := range courses {
		if courses[i].ID == id {
			return &courses[i]
		}
	}
	return nil
}
```

返回 &courses[i]，取得切片中实际元素的地址。找不到则返回 nil。

```go
coursePtr := findCoursePointerByID(courses, "1")
if coursePtr == nil {
	fmt.Println("课程不存在")
} else {
	coursePtr.Name = "通过指针修改的课程"
}
fmt.Println(courses)
```

不要误写为返回 range 副本的地址：

```go
for _, course := range courses {
	if course.ID == id {
		return &course
	}
}
```

这样的指针指向副本，修改它不会修改原切片元素。不是所有返回地址的写法都指向原数据。

当前仓库中指针查找函数与调用作为注释保留，需要同时恢复才能运行这个练习。

## 6 普通查找的对照

```go
course, found := findCourseByID(courses, "1")
if found {
	course.Name = "修改副本"
	fmt.Println("副本：", course)
}
fmt.Println("原切片：", courses)
```

普通查找返回 Course，因此副本改名、原切片不变是正常结果。

打印应放在赋值之后，否则观察到的是修改之前的值。

指针查找练习先采用“找到后立即修改”。持有元素指针期间，如果 append 重新分配底层数组，原指针仍指向旧数组；删除或排序也可能改变原位置中的元素。它不是永久绑定某个课程 ID 的引用。

## 7 方法：把行为关联到类型

普通函数：

```go
func completeCourse(course *Course) {
	course.Completed = true
}
```

对应的方法：

```go
func (c *Course) Complete() {
	c.Completed = true
}
```

函数名前面的 `(c *Course)` 叫接收者。c 表示本次调用所操作的课程指针，方法名是 Complete。

方法写在 struct 定义外、其他函数外，而不是放进结构体的大括号中。

调用：

```go
course := Course{
	ID:   "4",
	Name: "学习 Go 指针",
}
course.Complete()
fmt.Println(course.Completed) // true
```

对这里可取地址的变量，Go 允许自动取地址，上面的调用相当于：

```go
(&course).Complete()
```

方法让行为和类型组织在一起：course.Complete() 表达这门课程执行完成操作。不是每个结构体实例都保存了一份函数代码，方法定义是共享的。

## 8 值接收者与指针接收者

```go
func (c Course) CompleteCopy() {
	c.Completed = true
}
```

值接收者得到结构体副本，修改其 bool 字段不影响原课程。

```go
course := Course{Name: "接收者对比"}
course.CompleteCopy()
fmt.Println(course.Completed) // false
course.Complete()
fmt.Println(course.Completed) // true
```

需要修改原结构体时使用指针接收者。方法并没有改变 Go 的值传递规则，区别仍是复制结构体还是复制指针。

方法不仅让调用更直观，也是后续接口学习的基础。Go 根据类型具有的方法判断其是否满足接口，下一节再展开。

## 9 当前代码与练习验证

当前 main 创建了独立的 course 变量，调用 Complete，然后打印它。该变量没有追加到 courses 切片，因此方法修改这个变量不会让课程列表凭空增加一项。

当前实际运行的重点是指针接收者方法；普通值查找和指针查找的对照代码保留在注释中。本次检查不会把注释代码声称为已执行。

可以依次练习：

1. 比较普通函数接收 Course 与 *Course。
2. 分别恢复普通查找和指针查找，观察改名后原切片。
3. 对比值接收者与指针接收者的方法调用。
4. 查找不存在的 ID，先检查 nil 再访问字段。

验证使用 go run .、go vet ./... 和 go test ./...。已有测试主要覆盖课程查找、改名和删除；本节方法的简单赋值通过实际程序输出验证。
