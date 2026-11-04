# 第六节 JSON 编码与解码

上一篇：[第五节 接口与使用场景](./05-interfaces-and-when-to-use-them.md)。

本节把课程结构体、切片转换成 JSON，再把 JSON 还原为 Go 数据，为前后端交互做准备。

## 1 导入标准库与字段标签

```go
import "encoding/json"
```

encoding/json 是标准库，不需要安装第三方依赖。

```go
type Course struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}
```

反引号中的内容叫结构体标签。json 标签指定 JSON 字段名，Go 中仍然用 course.Name，JSON 中使用 name。

字段首字母要大写，才能被标准库正常导出。仅写 json 标签不能让小写的未导出字段参与普通编码。没有标签时，默认使用导出的 Go 字段名。

## 2 编码一门课程

```go
jsonCourse := Course{
	ID:        "200",
	Name:      "学习 Go JSON",
	Completed: false,
}

data, err := json.Marshal(jsonCourse)
if err != nil {
	fmt.Println("JSON 编码失败：", err)
	return
}
fmt.Println(string(data))
```

输出：

```json
{"id":"200","name":"学习 Go JSON","completed":false}
```

Marshal 返回 ([]byte, error)：成功的 JSON 内容是字节切片，错误通过 err 表达。

string(data) 只是将字节转换成字符串以便显示，不是再次编码。直接 fmt.Println(data) 通常打印字节数值列表。

这个功能与前端 JSON.stringify 的用途相近，但 Go 返回字节切片和错误。

## 3 编码课程切片

```go
listData, err := json.Marshal(courses)
if err != nil {
	fmt.Println("JSON 编码失败：", err)
	return
}
fmt.Println(string(listData))
```

结构体通常编码为对象，包含课程的切片编码为数组：

```json
[
  {"id":"1","name":"学习 Go 结构体","completed":false},
  {"id":"2","name":"学习 Go 切片","completed":false}
]
```

Marshal 默认输出紧凑格式，示例展开仅为了阅读。当前程序有三门课程，以实际切片内容为准。

注意 nil 切片默认编码为 null，非 nil 的空切片编码为 []；上面的“数组”对应当前已初始化的课程切片。

## 4 将 JSON 解码回切片

在编码之后继续写：

```go
var decodedCourses []Course

err = json.Unmarshal(listData, &decodedCourses)
if err != nil {
	fmt.Println("JSON 解码失败：", err)
	return
}

for _, course := range decodedCourses {
	fmt.Println(course.ID, course.Name, course.Completed)
}
```

Unmarshal 的两个参数：

| 参数 | 含义 |
| --- | --- |
| listData | 要解析的 JSON 字节数据 |
| &decodedCourses | 结果要写入的目标变量地址 |

Unmarshal 只返回 error，解析结果写入第二个参数指向的变量。

前面已声明 err，这里只有对它赋值，所以用 `=`，不能只写 `err := ...` 重复声明同一作用域下的变量。

## 5 为什么解码要传指针

```go
var decodedCourses []Course
json.Unmarshal(listData, &decodedCourses)
```

传地址后，函数才能更新 decodedCourses 这个变量本身。解码可能需要创建切片、增加长度并设置底层数组位置。

这与“传入切片后修改已有元素”不同：只复制切片描述信息，无法替换调用方的切片变量；传入指向切片的指针才可以写回完整结果。

对比两种 API：

```go
// 返回新切片描述，由调用方接住
courses = append(courses, newCourse)

// 传入目标地址，由函数写回
err = json.Unmarshal(listData, &decodedCourses)
```

直接传 decodedCourses 而不是地址，会返回参数错误。这里需要非 nil 的目标指针。

decodedCourses 本身是 nil 切片没有关系，&decodedCourses 是指向一个真实变量的非 nil 指针，解码器可以写入它。

## 6 解码手写 JSON 对象

以下是补充练习，不是当前 main 中已运行的代码：

```go
jsonText := `{"id":"10","name":"学习 JSON 解码","completed":true}`
var decodedCourse Course

err := json.Unmarshal([]byte(jsonText), &decodedCourse)
if err != nil {
	fmt.Println("解码失败：", err)
	return
}
fmt.Println(decodedCourse.Name)
fmt.Println(decodedCourse.Completed)
```

如果把它加到已有 err 的同一作用域，改用 `err = ...`。

反引号表示 Go 原始字符串，内部双引号无需转义。[]byte(jsonText) 把字符串转换为解码器需要的字节切片。

对象解码到 Course，数组解码到 []Course，目标类型要与数据结构对应。

## 7 解码失败与业务校验

把 completed 改成字符串：

```json
{"id":"10","name":"学习 JSON","completed":"是"}
```

目标字段是 bool，因此会返回类型错误。JSON 语法错误也会使解码失败。

错误发生时目标可能已经被部分填充，所以必须先检查 err，再将结果作为完整有效的数据使用。

解码成功不代表业务合法。例如名称为空的 JSON 仍可以成功解码。字段缺失也未必报错：解码到新声明的结构体时，未提供的字段保留零值。默认未知字段通常会被忽略。

课程名称非空、ID 是否存在等业务规则，仍需在解码之后校验。

## 8 与前端概念对应

| Go | 用途相近的 JavaScript 操作 |
| --- | --- |
| json.Marshal | JSON.stringify |
| json.Unmarshal | JSON.parse |
| json 标签 | 控制字段在 JSON 中的名称 |

Go 需要指定接收数据的目标类型并处理 error；不能认为与 JavaScript API 的返回形式完全一样。

## 9 本次检查与后续练习

当前 main 已完成 courses → JSON 字节 → decodedCourses 的往返过程。实际输出三门课程，ID、名称和完成状态正确。

格式检查、go vet 与现有测试通过。现有测试主要覆盖课程 CRUD，本节编码解码通过 go run . 的实际输出核对，尚未为 JSON 添加独立自动化测试。

练习方向：

1. 修改一门课程的 Completed，再观察编码与解码后的值。
2. 单独解码一门课程。
3. 输入类型错误的 JSON，确认错误处理分支。
4. 输入空名称，理解解码成功与业务校验成功的区别。

下一步可以用 net/http 返回课程 JSON，让浏览器通过接口读取课程列表。
