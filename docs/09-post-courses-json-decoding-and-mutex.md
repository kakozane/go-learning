# 第九节 POST 新增课程、JSON 请求体与互斥锁

上一篇：[第八节 路径参数、课程详情与关键词筛选](./08-path-parameters-and-course-search.md)。

本节增加 POST /courses，让客户端发送课程名称，后端校验后创建课程并返回 HTTP 201。仍使用内存切片，没有数据库。

## 1 新增接口的处理流程

```text
接收 POST /courses
  → 从请求体解码 JSON
  → 校验课程名称
  → 加锁
  → 生成 ID、添加课程、递增计数器
  → 返回 HTTP 201 和课程 JSON
  → 处理函数结束，释放锁
```

GET /courses 与 POST /courses 路径相同，但请求方法不同，因此分别注册处理函数。

## 2 用结构体接收请求

```go
type CreateCourseRequest struct {
	Name string `json:"name"`
}
```

客户端按 JSON 字段名发送：

```json
{"name":"学习 Go HTTP"}
```

ID 由后端生成，Completed 默认为 false，不要求客户端提供。

```go
var req CreateCourseRequest
```

这是 `var 变量名 类型` 的声明语法。没有显式初始值时使用零值，此时 req.Name 为 ""。当前代码选择另一种等价写法：

```go
req := CreateCourseRequest{}
```

结构体描述接收的数据形状，但不会自动执行“名称不能为空”等业务校验。

## 3 拆开理解 Decoder

```go
decoder := json.NewDecoder(r.Body)
err := decoder.Decode(&req)
if err != nil {
	http.Error(w, "请求体必须是合法的课程 JSON", http.StatusBadRequest)
	return
}
```

- r.Body 是可读取的请求体，不是已经读好的 []byte。
- NewDecoder 创建解码器，指定接下来从哪里读取。
- Decode 才开始读取 JSON，根据字段标签将解析出的值写入 req。
- &req 提供变量地址，让解码器能修改它。
- err 为 nil 表示这次解码成功；错误时必须先返回，不继续使用数据。

这些步骤可以合并成一句，当前保留拆开的写法方便理解：

```go
err := json.NewDecoder(r.Body).Decode(&req)
```

不能理解成“全部解析成功后才一次性赋值”：解码出错时，部分字段可能已经被写入。

## 4 Decoder 和 Unmarshal 的区别

| API | 输入 | 常见场景 |
| --- | --- | --- |
| json.Unmarshal(data, &req) | 已有 []byte | 解码之前拿到的 JSON 字节 |
| json.NewDecoder(reader).Decode(&req) | io.Reader | 从请求体或文件读取 JSON |

也可以先用 io.ReadAll(r.Body) 读取字节，再调用 Unmarshal，但需要分别检查读取和解码的错误。当前直接用 Decoder，省去手动读取这一步。

Unmarshal 检查传入的整段 JSON 数据；一次 Decode 只解码一个 JSON 值，不保证后面没有额外内容。当前教学接口尚未添加第二次读取并检查 EOF，也没有限制请求体大小，不能把它当成已经完成所有输入限制的生产接口。

## 5 JSON 必须与结构体完全一样吗

对于刚声明的 CreateCourseRequest：

| JSON | 默认结果 |
| --- | --- |
| {"name":"学习 Go"} | 成功，Name 得到对应字符串 |
| {} | 解码成功，Name 保持空字符串 |
| {"name":"学习 Go","age":18} | 默认忽略未知的 age 字段 |
| {"name":123} | 字段类型不匹配，返回错误 |
| [{"name":"学习 Go"}] | 数组不能解码到该结构体，返回错误 |

字段缺失通常保留目标已有值；这里目标刚创建，所以是零值。JSON null 对非指针字符串字段不会自动变成字符串，当前新变量仍为空，会被名称校验拒绝。

若希望未知字段报错，可在 Decode 前调用 decoder.DisallowUnknownFields()，但当前代码没有启用。

前后端需要约定 JSON 字段名、类型，以及不能为空等业务规则。解码成功不代表业务合法。

## 6 名称校验

```go
name := strings.TrimSpace(req.Name)
if name == "" {
	http.Error(w, "课程名称不能为空", http.StatusBadRequest)
	return
}
```

全空格和缺失名称都会被拒绝，返回 HTTP 400。保存时使用处理后的 name，所以名称两端的空白不会进入课程切片。

## 7 为什么 nextID 是整数

main 在初始三门课程之后声明：

```go
nextID := 4
var coursesMu sync.Mutex
```

nextID 用于递增计数，Course.ID 沿用 string 类型，所以创建时转换：

```go
course := Course{
	ID:        strconv.Itoa(nextID),
	Name:      name,
	Completed: false,
}
courses = append(courses, course)
nextID++
```

strconv 是标准库包，Itoa 将 int 转成数字的字符串形式：

```go
strconv.Itoa(65) // "65"
string(65)      // "A"，整数对应的 Unicode 字符
```

不要把 string(整数) 当成十进制数字转字符串。之前的 string(data) 中 data 是 []byte，是另一种转换。

字符串不能自增，"4" + "1" 得到 "41"。因此 nextID 使用 int，输出课程 ID 时转换成 string。变量名本身不决定类型。

从 4 开始依赖当前初始 ID 是 1、2、3；将来修改初始数据或接入数据库，需要一起调整 ID 策略。

## 8 互斥锁保护什么

HTTP 请求可能并发执行。courses 和 nextID 定义在处理函数外，是多个请求共享的数据。

```go
coursesMu.Lock()
defer coursesMu.Unlock()
```

Lock 取得锁；其他尝试取得同一把锁的请求需要等待。defer 让 Unlock 在当前处理函数返回时执行，包括提前 return 的情况。

所有访问共享课程的 HTTP 路径都要遵守同一把锁。当前两个 GET 接口和 POST 的数据修改部分都使用 coursesMu。

生成 ID、append、nextID++ 必须位于同一段加锁范围内。否则两个新增请求可能读到相同的 nextID，或者并发修改切片。

锁要在 main 中创建一次，不能在每个处理函数里各自新建一把；互不相同的锁无法保护同一份数据。

## 9 为什么 GET 开头加锁，POST 中间加锁

决定位置的是“何时访问共享数据”，不是 GET 或 POST 这个名称。

GET 很快就要读取 courses，因此示例直接在开头加锁。读取关键词本身不访问共享课程，也可以放到锁外。

POST 前半段只处理当前请求独有的 req、name，不需要课程锁。读取请求体可能比较慢，先完成解码和校验，再取得锁，避免其他课程请求等待不必要的工作。

当前使用 sync.Mutex，读和写都会排队，并不允许多个读取者同时进入。为方便学习，defer Unlock 放在处理函数范围，连 JSON 编码和响应写入也在锁内；慢响应可能阻塞其他课程请求，后续需要吞吐量时再缩短加锁范围。

## 10 w 的三个方法

http.ResponseWriter 接口本身包含三个方法：

| 方法 | 作用 |
| --- | --- |
| Header() | 取得响应头集合 |
| WriteHeader(statusCode) | 提交状态码和当前响应头 |
| Write(data) | 写入响应体字节，返回字节数和错误 |

Header().Set 可以拆成：

```go
headers := w.Header()
headers.Set("Content-Type", "application/json")
```

Set 是响应头集合的方法，不是直接定义在 ResponseWriter 上的方法。

当前新增接口的顺序：

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败：", err)
}
```

先准备响应头，再提交 201，最后写 JSON 字节。省略 WriteHeader 时，第一次 Write 默认提交 200；之后再修改最终状态码就晚了。

## 11 http.Error 为什么需要 w

```go
http.Error(w, "课程名称不能为空", http.StatusBadRequest)
return
```

w 指向当前请求的响应写入器，http.Error 通过它设置文本响应头、状态码和错误文字，告诉当前客户端发生了什么。

它不是 w 的方法，也不是打印日志：fmt.Println 输出到服务端终端，http.Error 输出给客户端。http.Error 返回后，外层处理函数仍会继续，因此需要紧接着 return。

其他接收 w 的辅助函数还有 http.Redirect、fmt.Fprintln，以及 json.NewEncoder(w).Encode(data)。当前仍使用 Marshal 加 Write，不急着替换。

## 12 HTTP 201 与业务 code

201 Created 表达请求成功并创建了新资源，适合本节新增课程。并不是所有 POST 都要返回 201，例如登录并不一定创建资源。

有些团队约定所有成功请求都返回 200，也能使用；本项目用 200 表达查询成功、201 表达创建成功、400 表达参数错误、404 表达资源不存在。

```go
w.WriteHeader(http.StatusCreated) // 真正的 HTTP 状态码
```

响应结构的 Code: 201 只是 JSON 业务字段，两者不会自动同步。如果项目连业务失败都返回 HTTP 200，前端只判断 response.ok 就不够，还需要检查业务 code。

## 13 数据存在哪里

新增课程存入服务进程内存中的 courses。刷新浏览器不会清空，后端重启后会重新初始化为三门课程，nextID 也回到 4。

目前没有数据库、文件持久化或多实例间共享存储。客户端收到 201 表示这次内存创建完成，不代表已经写入数据库。

## 14 实际练习

启动服务后，在另一个终端运行：

```sh
curl -i http://localhost:8080/courses \
  -H 'Content-Type: application/json' \
  -d '{"name":"学习 Go HTTP"}'
```

curl 的 -d 默认使用 POST。预期状态 201，data 是新课程对象，Completed 为 false。再查询 /courses 和 /courses/4 应能看到新增课程。

继续练习：

1. 将 name 改成 "  学习 Go HTTP  "，确认保存的是去掉两端空白后的名称。
2. 发送空名称或 {}，确认 400 且没有新增。
3. 发送错误 JSON 或数字类型的 name，确认 400。
4. 连续新增两门课程，确认 ID 递增。
5. 重启后重新查询，理解内存数据与持久化的区别。

## 15 本次验证与范围

提交前 gofmt、go vet 和编译检查通过。go test 显示 [no test files]，当前没有持久化的自动化测试文件。

使用当前 main.go 的临时副本，仅替换监听地址为临时本机端口，以 go build -race 构建并运行接口检查，没有修改或打断原来运行的服务。验证了：

- 非法 JSON、字段类型错误、空名称、缺失名称、null、数组和空请求体返回 400，列表仍为原来的三门课程。
- 正常创建返回 201，名称被去掉两端空白，ID 为 4，完成状态为 false，详情接口能够读到它。
- 交错并发执行 20 次查询与 20 次新增，新增 ID 不重复，最终列表数量正确。
- 不存在的课程详情返回 404。
- 上述请求运行期间，竞态检测器没有报告数据竞争；这不等于证明所有未来代码路径都没有竞态。

没有进行前端联调、数据库持久化或生产压测。当前仍允许未知 JSON 字段，单次 Decode 未检查尾随内容，未设置请求体大小上限；锁仍覆盖响应写入。这些是当前本地学习示例的明确范围，后续可分别练习完善。
