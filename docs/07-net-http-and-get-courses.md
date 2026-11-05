# 第七节 net/http 与 GET 课程列表接口

上一篇：[第六节 JSON 编码与解码](./06-json-encoding-and-decoding.md)。

本节将课程数据通过 HTTP 返回给浏览器。当前代码使用 Go 标准库，不需要安装框架；课程保存在内存中，还没有连接数据库。

## 1 从终端程序变成 HTTP 服务

以前运行 main，打印结果后程序就结束了。现在 main 创建路由并启动服务，等待浏览器、JavaScript 或其他客户端发来请求。

```text
go run .
  → 创建课程切片
  → 创建 mux 并注册处理函数
  → ListenAndServe 监听 localhost:8080
  → 客户端请求 GET /courses
  → mux 找到对应处理函数
  → 课程编码为 JSON，写入 HTTP 响应
  → 客户端收到响应，服务继续等待其他请求
```

不是不断重新执行 main。每次匹配的请求都会调用处理函数；处理函数 return 只结束本次处理，不会停止服务。

## 2 mux 是什么

mux 是 multiplexer 的缩写，意思是多路复用器。这里可以理解为请求分发器。

```go
mux := http.NewServeMux()
```

mux 只是变量名，也可以叫 router。它负责根据请求方法和路径选择处理函数，监听主机和端口的是 HTTP 服务。

一个 mux 可以注册很多接口。例如下面是扩展示例，不是当前已实现的接口：

```go
mux.HandleFunc("GET /courses", listCourses)
mux.HandleFunc("GET /users", listUsers)
```

普通学习项目使用一个 mux 就够了，但 Go 不限制只能创建一个 mux；mux 本身也不绑定某个端口，可以被不同服务复用。

## 3 注册处理函数

```go
mux.HandleFunc("GET /courses", func(w http.ResponseWriter, r *http.Request) {
	// 收到匹配请求时才执行
})
```

HandleFunc 的第一个参数是匹配规则，第二个参数是处理函数。这里传入的是匿名函数。

`GET /courses` 表示方法为 GET、路径为 /courses；Go 的这类 GET 规则也匹配 HEAD 请求。带方法的规则写法要求 Go 1.22 或更新版本，项目版本满足要求。

注册时只是建立对应关系，不会立刻运行函数体。函数引用了外部的 courses，因此可以返回 main 中的课程切片。

## 4 w 和 r 分别是什么

```go
func(w http.ResponseWriter, r *http.Request)
```

| 参数 | 含义 | 用途示例 |
| --- | --- | --- |
| w | 响应写入器 | 设置响应头、状态码、响应体 |
| r | 本次请求 | 读取方法、URL、查询参数、请求体 |

以下是读取请求的示例：

```go
fmt.Println(r.Method)
fmt.Println(r.URL.Path)
fmt.Println(r.URL.Query().Get("name"))
```

当前接口总是返回固定课程列表，不需要读取 r，不代表它不能处理请求数据。函数参数暂时没使用也允许编译。

http.Request 是结构体，*http.Request 是指向它的指针。传指针可以共享请求对象、避免复制完整结构体；指针不只用于修改数据，也可以用于读取。

http.ResponseWriter 是接口类型，实际对象由 HTTP 服务提供。接口背后的实现可以是指针，所以不能根据 w 的声明没有星号，就认为无法通过它写入响应。这里也用到了前面学过的“通过接口调用方法”。

## 5 当前接口的响应结构

```go
type ResData struct {
	Code    int      `json:"code"`
	Data    []Course `json:"data"`
	Message string   `json:"message"`
}
```

构造响应并编码：

```go
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
```

Marshal 返回 JSON 字节切片和错误。先检查错误，再使用结果。

http.Error 写入错误响应，这里状态码是 500。后面的 return 防止继续执行成功响应代码。这个错误响应是文本，不会自动使用 ResData 的 JSON 结构。

成功响应的 data 是课程数组：

```json
{
  "code": 200,
  "data": [
    {"id": "1", "name": "学习 Go 结构体", "completed": false},
    {"id": "2", "name": "学习 Go 切片", "completed": false},
    {"id": "3", "name": "学习 Go 函数", "completed": false}
  ],
  "message": "请求成功！"
}
```

## 6 响应头和响应体

```go
w.Header().Set("Content-Type", "application/json")

_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败：", err)
}
```

Content-Type 告诉客户端这是 JSON，不负责将结构体转换为 JSON。转换已经由 Marshal 完成。

Write 将字节写入响应，返回“写入的字节数、错误”。这里用下划线忽略字节数，保留错误检查。客户端断开连接等情况可能导致写入失败。

响应头需要在 Write 之前设置。如果没有先调用 WriteHeader，第一次 Write 会隐式使用 HTTP 200。

**JSON 中的 code 是业务字段，HTTP 状态码是协议层的状态，两者不会自动同步。** 即使把 ResData.Code 改成 500，也不会自动让 HTTP 状态变成 500。

写入失败时记录日志即可，不再尝试追加另一段错误响应，因为响应可能已经部分发出。

## 7 两种错误处理写法

当前写法复用了处理函数中 Marshal 声明的 err：

```go
_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败：", err)
}
```

也可以将声明放到 if 中：

```go
if _, err := w.Write(data); err != nil {
	log.Println("响应写入失败：", err)
}
```

第二种的分号分隔“初始化语句”和“判断条件”。这个新 err 只在该 if 及其 else 范围内有效，会遮蔽外层同名变量。

两种写法都可以，不需要为了简短强行改写。当前 err 在每次处理请求时分别创建，不是多个请求共享的全局变量。

## 8 启动并保持服务运行

```go
fmt.Println("准备启动服务：http://localhost:8080")
err := http.ListenAndServe("localhost:8080", mux)
if err != nil {
	log.Fatal("服务启动失败：", err)
}
```

localhost 是本机地址，8080 是端口，mux 是接收请求后负责分发的处理器。

ListenAndServe 在服务运行时持续阻塞，不会立刻执行后面的代码。返回错误时，log.Fatal 打印错误并退出程序。例如端口被占用可能导致启动失败。

“准备启动服务”的打印只说明即将尝试启动，并不保证已经成功监听。查看是否有后续错误，并实际请求接口才能进一步确认。

在运行服务的终端按 Ctrl+C 停止。普通 go run 不自动热更新，修改代码后需要停止并重新运行。

## 9 为什么 fmt.Println 没有在控制台出现

要区分两处控制台：

- 后端终端：运行 go run . 的终端，接收 Go 的 fmt.Println 和 log 输出。
- 浏览器开发者工具 Console：显示浏览器 JavaScript 的 console.log。

Go 的打印不会自动出现在浏览器 Console。并且放在处理函数里的打印，只有请求匹配该接口时才会执行。

```go
fmt.Println(resData)      // Go 结构体的展示形式
fmt.Println(string(data)) // JSON 文本，当前代码采用这一种
```

排查顺序：确认查看的是后端终端 → 确认修改后重启 → 请求 /courses → 检查是否请求到了正确端口和服务。

## 10 string 和 strings

| 名称 | 含义 | 例子 |
| --- | --- | --- |
| string | 内置字符串类型，也用于类型转换 | var name string、string(data) |
| strings | 字符串处理标准库包，需要导入 | strings.TrimSpace(name) |

string(data) 把 JSON 字节切片转换成字符串，便于打印；strings.TrimSpace(name) 返回去掉两端空白的字符串。它们不是同一个东西。

## 11 手动验证步骤

在项目目录运行：

```sh
go run .
```

保持服务运行，在浏览器打开 http://localhost:8080/courses，或者在另一个终端执行：

```sh
curl -i http://localhost:8080/courses
curl -i http://localhost:8080/missing
curl -i -X POST http://localhost:8080/courses
```

预期分别是：课程列表的 HTTP 200 JSON 响应、未匹配路径的 404、不支持方法的 405。

访问一次课程接口，后端终端应打印一次 JSON。浏览器自动发出的其他路径请求，不会执行课程处理函数。

## 12 当前练习范围

当前仅实现读取内存课程列表，没有 HTTP 新增、修改、删除，也没有数据库持久化或前端联调。旧的 CRUD 函数及学习注释不等于已经注册了对应的 HTTP 接口。

HTTP 服务可以并发处理请求。当前只有读取；将来增加修改共享切片的接口时，需要处理并发访问问题。前端跨端口请求可能涉及代理或 CORS，留到联调时学习。

当前旧 main_test.go 已删除，go test 若显示 [no test files]，仅表示没有测试用例，不能说业务测试全部通过。本节通过编译、静态检查和实际接口请求核对行为。
