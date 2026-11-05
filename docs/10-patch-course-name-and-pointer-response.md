# 第十节 PATCH 修改课程名称与指针响应

上一篇：[第九节 POST 新增课程、JSON 请求体与互斥锁](./09-post-courses-json-decoding-and-mutex.md)。

本节实现 PATCH /courses/{id}，修改已有课程的名称，并返回修改后的课程。课程仍保存在服务进程内存中。

## 1 PATCH 不限制修改字段数量

PATCH 描述对资源的修改，可以修改一个字段，也可以修改多个甚至所有允许修改的字段。重点是接口的修改语义，不是字段数量。

PUT 通常表达期望的完整资源状态，用于整体替换；不能默认认为省略字段会保持原值。具体字段规则仍以接口约定为准。

本次接口只允许修改名称，课程 ID 和完成状态不变。名称是本次请求的必填字段，省略或为空都会返回 400；当前并不是支持所有字段可选的通用 PATCH 接口。

## 2 处理流程

```text
取得路径中的 id
  → 解码 JSON 到临时变量 req
  → 去掉名称两端空白，拒绝空名称
  → 取得课程锁
  → 根据 ID 找到课程指针
  → 不存在则返回 404
  → 修改课程名称
  → 编码响应，返回 HTTP 200
  → 函数退出，释放锁
```

路由注册在 ListenAndServe 之前：

```go
mux.HandleFunc("PATCH /courses/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 读取、校验、查找、修改、响应
})
```

## 3 为什么先解码到 req

当前代码用 Course 接收请求：

```go
var req Course
decoder := json.NewDecoder(r.Body)
err := decoder.Decode(&req)
if err != nil {
	http.Error(w, "解码失败", http.StatusBadRequest)
	return
}
```

这里示例使用小写 decoder；当前练习文件中的局部变量名是 Decoder，命名不同不影响行为。

r.Body 是原始请求体的数据流，req 是接收解码结果的 Go 结构体。不能把请求体直接赋值给字符串字段，需要解析 JSON 才知道 name 对应的字符串是什么。

```go
name := strings.TrimSpace(req.Name)
if name == "" {
	http.Error(w, "课程名称不能为空", http.StatusBadRequest)
	return
}
```

req 相当于一个临时接收和检查数据的位置。解码与业务校验成功后，才修改真正的课程。

技术上可以直接 Decode(currentCourse)，但不适合这里：解码失败时部分字段可能已经改变；空名称校验前就可能写入；客户端还可能改变 ID 或 Completed；读取请求体期间也需要持有共享数据锁。

当前虽然用 Course 接收数据，实际只赋值 Name，其他字段不会被写回。未来可以定义只含 Name 的修改请求结构体，明确接口允许修改的字段，当前不强行重构。

## 4 查找函数返回指针

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

&courses[i] 是切片中实际元素的地址，不是 range 临时副本的地址。

```go
coursesMu.Lock()
defer coursesMu.Unlock()

currentCourse := findCoursePointerByID(courses, id)
if currentCourse == nil {
	http.Error(w, "课程不存在", http.StatusNotFound)
	return
}

currentCourse.Name = name
```

currentCourse 的类型是 *Course，保存指向实际课程的指针。必须先判断 nil，再访问字段。

查找、修改、读取修改后的课程值都在同一把锁保护下，避免与其他课程请求并发读写。不能把这个指针随意交给锁外的异步任务使用。

## 5 currentCourse 与 *currentCourse

| 表达式 | 含义 | 类型 |
| --- | --- | --- |
| currentCourse | 指向课程的指针 | *Course |
| *currentCourse | 解引用，取得它指向的课程值 | Course |
| &courses[i] | 取得切片元素的地址 | *Course |

访问结构体字段时，Go 允许自动解引用：

```go
currentCourse.Name = name
// 等价于：
(*currentCourse).Name = name
```

因此这里修改的是原切片里的课程。

响应结构的 Data 定义为 Course，所以需要传入课程值：

```go
resData := FindCourseResData{
	Code:    200,
	Data:    *currentCourse,
	Message: "修改成功",
}
```

Data: *currentCourse 将该课程值复制到响应结构中。currentCourse 本身仍然是指针，星号没有改变变量的类型，只是执行取值操作。

## 6 http.Error 之后为什么要 return

http.Error 写入错误响应，不会替调用者退出处理函数。

```go
if err != nil {
	http.Error(w, "解码失败", http.StatusBadRequest)
	return
}
```

如果不 return，后面的课程修改仍会执行。之前的练习版本发送数字类型 name 时返回 400，却继续把课程名称清空，还将成功 JSON 拼接到错误文字后面。

nil 分支同样必须 return，否则接下来的 currentCourse.Name 会触发空指针 panic。标准 HTTP 服务通常会恢复处理函数的 panic 并关闭相关连接，但这不是正常的 404 响应，不能依赖它代替业务错误处理。

当前请求参数错误返回 400，指定课程不存在返回 404，两种含义不同。

## 7 先编码，再提交成功响应

```go
data, err := json.Marshal(resData)
if err != nil {
	http.Error(w, "课程数据编码失败", http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)

_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败", err)
}
```

当前文件使用 WriteHeader(200)，与 http.StatusOK 等价。

先检查 Marshal 的错误，不能直接用下一句赋值覆盖 err。先完成编码再写状态码，编码失败时才有机会返回 500。当前响应字段都是可编码的基本类型，错误分支仍保留以保持清晰的错误处理顺序。

响应写入失败时记录日志，不再次调用 http.Error：成功状态可能已经发出，响应也可能已部分写入，不能再拼接另一种错误响应。网络写入失败也不等于客户端参数错误。

## 8 末尾的错误日志需要 return 吗

```go
_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败", err)
}
// 当前函数到这里结束
```

后面没有代码时，可以省略 return，函数自然结束。若后面还有不应该继续执行的工作，就需要 return。

判断依据不是“遇到 err 必须写 return”，而是“错误后是否需要跳过后续代码”。提前 return 和正常执行到末尾都会执行已登记的 defer Unlock。

## 9 curl 练习

修改代码后重启服务，在另一个终端执行：

```sh
curl -i -X PATCH http://localhost:8080/courses/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"学习 Go 修改接口"}'
```

预期 HTTP 200，data 返回修改后的课程。再次 GET /courses/1，可以确认内存中的课程已经改变。

测试空名称，预期 400，原课程保持不变：

```sh
curl -i -X PATCH http://localhost:8080/courses/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"   "}'
```

测试不存在的课程，预期 404：

```sh
curl -i -X PATCH http://localhost:8080/courses/999 \
  -H 'Content-Type: application/json' \
  -d '{"name":"测试课程"}'
```

还可以发送 {}、错误 JSON、数字类型 name，观察错误响应后原课程名称是否保持不变。

## 10 当前边界

仍沿用第九节的教学范围：数据只在内存中；一次 Decode 未检查后续额外 JSON；未限制请求体大小；锁覆盖响应写入。它是修改名称练习，不是已经完成全部生产输入限制的接口。

## 11 本次验证

格式检查、go vet 和编译通过；go test 显示 [no test files]，当前没有持久化的自动化测试文件。

使用当前源码的临时副本，仅替换监听地址为临时本机端口，并通过 go build -race 构建，实际验证了：

- 正常修改返回 200，去掉名称两端空白；请求附带其他 ID 和 Completed 值时，不改变原课程的这两个字段。
- 空名称、缺失名称、类型错误、非法 JSON、null、数组和空请求体返回 400，原课程保持不变。
- 不存在的课程返回 404，且不增加课程。
- 交错执行 30 次 GET、POST、PATCH 请求，响应状态正常，最终课程数量和 ID 唯一性正确。
- 本次请求执行期间没有竞态或 panic 报告，不代表证明所有未来代码路径都安全。

检查未打断已有服务，未进行 React 前端联调或数据库验证。
