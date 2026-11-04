# 第五节 接口与使用场景

上一篇：[第四节 指针、方法与接收者](./04-pointers-and-methods.md)。

本节通过课程与任务的完成操作学习接口，并区分语法练习和真正需要抽象的业务场景。

## 1 接口声明能力

```go
type Completer interface {
	Complete()
}
```

接口规定方法的名字、参数和返回值，不写实现。这里要求一个无参数、无返回值的 Complete 方法。

满足要求的类型可以作为 Completer 使用，Go 不需要显式写 implements。

## 2 两种类型分别实现方法

```go
type Course struct {
	ID        string
	Name      string
	Completed bool
}

func (c *Course) Complete() {
	c.Completed = true
}

type Task struct {
	Title string
	Done  bool
}

func (t *Task) Complete() {
	t.Done = true
}
```

*Course 和 *Task 都具有符合要求的方法，因此都满足 Completer。课程修改 Completed，任务修改 Done，各自决定怎么实现。

## 3 统一调用

```go
func completeAll(items []Completer) {
	for _, item := range items {
		item.Complete()
	}
}
```

[]Completer 是接口切片，元素可以保存满足同一接口的不同具体类型值。

```go
demoCourse := Course{ID: "100", Name: "学习 Go 接口"}
demoTask := Task{Title: "整理学习笔记"}

fmt.Println("完成前：", demoCourse.Completed, demoTask.Done)
completeAll([]Completer{&demoCourse, &demoTask})
fmt.Println("完成后：", demoCourse.Completed, demoTask.Done)
```

输出：

```text
完成前： false false
完成后： true true
```

第一轮调用 *Course 的 Complete，第二轮调用 *Task 的 Complete。循环不需要判断具体类型，也不需要知道各自的字段名称。

## 4 为什么传地址

方法的接收者是 *Course 与 *Task，所以满足接口的是这两个指针类型。

```go
completeAll([]Completer{&demoCourse, &demoTask})
```

直接将 demoCourse 放入这个接口切片会编译失败。虽然对可取地址的变量，demoCourse.Complete() 可以自动取地址调用，但赋给接口时不会自动进行这种转换。

临时把 Task 的方法改名为 Finish 后，*Task 也不再满足 Completer，因为方法名不符合约定。方法签名也必须匹配。

## 5 同名方法与函数重载

不同接收者类型可以拥有同名方法：

```go
func (c *Course) Complete() {}
func (t *Task) Complete() {}
```

同一个包中的普通函数不能重名，即使参数不同：

```go
// 错误示意，不能一起编译
func complete(course Course) {}
func complete(task Task) {}
```

同一个基础类型也不能分别给值接收者和指针接收者定义同名方法：

```go
// 错误示意，不能一起编译
func (c Course) Complete() {}
func (c *Course) Complete() {}
```

Go 不通过普通函数参数类型区分同名函数。接口利用的是不同类型具有相同方法签名的能力。

## 6 为什么不直接分别调用

在当前只有两个已知对象的场景，直接调用完全合理，而且更简单：

```go
demoCourse.Complete()
demoTask.Complete()
```

定义接口、构造切片和循环并没有让这两行代码更短。本例用于观察接口机制，不代表所有同名方法都应包装成接口。

同理，只有下面这一层转发时，通常不必为了抽象而增加它：

```go
func completeItem(item Completer) {
	item.Complete()
}
```

当前仓库保留它作为语法练习。

## 7 接口更有价值的情况

当处理流程确定，但具体实现需要替换时，接口可以让流程不依赖某一种实现。

例如课程存储的概念示例：

```go
type CourseStore interface {
	Save(course Course) error
}

func createCourse(store CourseStore, course Course) error {
	if strings.TrimSpace(course.Name) == "" {
		return errors.New("课程名称不能为空")
	}
	return store.Save(course)
}
```

这里需要 errors 和 strings 标准库。示例只展示校验与保存的分工，没有实现数据库或内存存储，也没有加入当前程序。

未来数据库存储和内存存储只要实现 Save(Course) error，都可以传入 createCourse。校验流程无需因具体存储实现而改变，测试也可以使用不依赖数据库的实现。

若实际只有一种实现、无需替换，先使用具体类型即可，不为可能发生的需求提前设计接口。

## 8 接入已有接口

一些标准库 API 本身要求接口参数，例如 fmt.Fprintln 的输出目标要满足 io.Writer。

```go
var buffer bytes.Buffer
fmt.Fprintln(&buffer, "你好")
fmt.Print(buffer.String())
```

此示例需要导入 bytes 和 fmt。*bytes.Buffer 已实现 io.Writer 要求的 Write 方法，所以可以传入。文件等其他实现也可以成为输出目标。

这里通常不需要自己重新定义接口，只需提供满足现有接口的类型。

## 9 选择依据

| 情况 | 选择 |
| --- | --- |
| 两个已知对象分别执行操作 | 直接调用各自方法 |
| 逻辑只依赖一个具体类型 | 优先使用具体类型 |
| 同一流程需要支持不同实现 | 考虑小而明确的接口 |
| 使用的库要求接口参数 | 提供满足接口的具体值 |

接口的价值主要是统一能力约定和隔离实现依赖，不保证减少代码行数。当前不需要强行把课程 CRUD 全部改成接口。

## 10 当前验证与下一步

当前 main 的接口练习实际运行后，课程和任务都从 false 变为 true。Go 格式检查、go vet 与现有测试用于检查当前代码；现有自动化测试主要覆盖课程 CRUD，接口示例通过实际运行验证。

下一步学习 JSON 编码、解码和结构体标签，让 Go 课程数据转换成前端能接收的格式，再进入 HTTP 接口开发。
