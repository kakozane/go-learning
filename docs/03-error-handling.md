# 第三节 error 错误处理

上一篇：[第二节 课程查找、修改与删除](./02-course-crud-and-slice-deletion.md)。

本节给修改课程名称增加校验，从“是否成功”进一步表达“为什么失败”。

## 1 bool 与 error 的区别

bool 表示是或否，适合 findCourseByID 中是否找到的判断。error 可以携带失败原因；返回 nil 通常表示没有错误。

| 返回结果 | 含义 |
| --- | --- |
| false | 失败或未找到，具体含义由函数约定 |
| errors.New("课程名称不能为空") | 名称不合法 |
| errors.New("课程不存在") | 没有匹配的课程 |
| nil（作为 error 返回值） | 操作成功 |

error 是 Go 内置类型；errors 是标准库包，errors.New 用来创建错误值。

它不会自动寻找错误原因：函数负责检测问题并返回原因，调用方负责处理。

## 2 当前函数实现

```go
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
```

使用时需要导入 errors 和 strings。

strings.TrimSpace 去掉两端空白，类似 JavaScript 的 trim。校验后必须把处理后的 name 保存进去，使用原始 newName 会把空格也保存下来。

本函数先校验名称，再查找 ID：两者都不合法时，首先返回名称错误。验证失败不会修改课程。

## 3 调用方要检查并使用错误

```go
err, index := updateNameByID(courses, "1", "")
if err != nil {
	fmt.Println("修改失败：", err)
} else {
	fmt.Println("修改成功：", courses[index].Name)
}
```

若只打印“修改失败”，就丢失了具体原因，看起来与 bool 没有区别。

错误是普通返回值，不是 JavaScript 的 throw。返回错误会结束当前函数，但不会自动中断调用方；调用方需要检查 err。

只有成功时才使用返回下标，避免用失败时的 -1 访问切片。

## 4 返回值顺序惯例

当前练习使用 `(error, int)`，语法合法且调用方对应正确。Go 通常将 error 放在最后，更常见的是：

```go
func updateNameByID(courses []Course, id string, newName string) (int, error)
```

对应成功返回 i, nil，失败返回 -1, err，调用方用 index, err 接收。如果调整顺序，函数、调用处和测试必须一起改，不能只改声明。

不需要将所有 bool 返回值替换为 error。单纯查找是否存在可以继续用 bool，需要表达操作失败原因时使用 error。

## 5 本节发现的问题

- 只打印固定提示，未展示 err：修改为同时打印错误值。
- 使用 TrimSpace 校验后仍保存 newName：改为保存 name。
- 测试仍把返回值当 bool，使用 !updated：error 不能这样判断，要使用 err != nil 或 err == nil。

测试由助手同步维护，已覆盖成功去空白、空字符串、空白名称、课程不存在，以及校验失败不改变数据。测试实现不是本节必须掌握的内容。

## 6 练习

分别传入有效名称、空字符串、仅空白名称和不存在的 ID。检查输出的错误原因是否准确，失败后原课程是否保持不变。

本节错误处理与测试已在此前提交验证；当前 main 中这段演示可能被注释，以便运行后续指针练习。
