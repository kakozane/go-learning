# 第十二节 修改完成状态与结构体指针字段

上一篇：[第十一节 DELETE 删除课程、切片赋值与 204 响应](./11-delete-courses-slices-and-no-content.md)。

本节实现 PATCH /courses/{id}/completed，将课程设置为已完成或未完成。请求必须提供布尔值 completed，不能把“没传”默认当成 false。

## 1 普通字段和指针字段怎么选

普通类型具有零值：bool 是 false，int 是 0，string 是空字符串，指针是 nil。

实际存储的课程使用普通 bool：

```go
type Course struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}
```

课程一定有完成状态，false 表示未完成，已经足够表达业务含义。

新增请求的名称也可以用 string：不传名称和传空名称都会被拒绝，无需区分。

但修改完成状态时，false 是合法值。如果请求字段也用 bool，解码到新变量后，{} 和 {"completed":false} 的结果相同，就无法判断客户端是否提供了状态。

因此本次请求结构体使用指针字段：

```go
type UpdateCourseCompletedRequest struct {
	Completed *bool `json:"completed"`
}
```

| JSON 请求 | 解码到新变量后的字段 |
| --- | --- |
| {} | nil |
| {"completed":null} | nil |
| {"completed":false} | 非 nil 指针，指向 false |
| {"completed":true} | 非 nil 指针，指向 true |

普通指针字段不能单独区分“没传”和“传 null”；本接口把两者都作为缺少有效状态处理。

## 2 其他适合指针字段的例子

| 字段 | 未提供的含义 | 提供零值的含义 | 可用类型 |
| --- | --- | --- | --- |
| 商品价格 | 不修改 | 0 表示免费 | *int |
| 个人简介 | 不修改 | 空字符串表示清空 | *string |
| 完成状态 | 拒绝请求或不修改 | false 表示未完成 | *bool |

这些是设计示例，不是当前项目实现。是否允许零值、未提供时如何处理，由接口约定决定。

同一个业务字段在存储结构体中用普通类型，在请求结构体中用指针类型很正常，它们的职责不同。不要把所有字段都改成指针。

## 3 结构体与对象属性的类比

```go
type UpdateCourseCompletedRequest struct {
	Completed *bool `json:"completed"`
}
```

这是定义一种类型，规定字段名称和类型。

```go
req := UpdateCourseCompletedRequest{}
```

这是创建该类型的值，存入 req。req.Completed 是字段访问，类似 JavaScript 对象的属性访问，但 Go 结构体字段是预先确定的，不能随意增加不存在的字段。

```text
req.Completed         字段访问
 decoder.Decode(&req)  方法调用
```

JSON 使用标签中的 completed，Go 代码使用字段名 Completed。

## 4 解码错误与 nil 都要检查

当前接口按以下顺序处理：

```go
req := UpdateCourseCompletedRequest{}
decoder := json.NewDecoder(r.Body)
err := decoder.Decode(&req)
if err != nil {
	http.Error(w, "请求数据格式错误", http.StatusBadRequest)
	return
}

if req.Completed == nil {
	http.Error(w, "必须提供完成状态", http.StatusBadRequest)
	return
}
```

检查解码错误是为了拒绝非法 JSON 或类型不匹配；检查 nil 是为了拒绝缺失或 null 字段。这两个检查不能互相替代。

之前练习中出现过的问题：

- 定义了请求结构体，却仍然解码到 Course，缺失状态被当成 false。
- 忽略 Decode 的错误，传字符串状态也返回成功。
- 改用 *bool 后没有判断 nil，发送 {} 或 null 导致解引用 panic。

解码失败时目标可能已部分赋值，因此即使指针不为 nil，也不能跳过 err 检查。错误响应后必须 return，避免继续修改数据。

## 5 加锁、查找，再修改

```go
coursesMu.Lock()
defer coursesMu.Unlock()

c := findCoursePointerByID(courses, id)
if c == nil {
	http.Error(w, "课程不存在！", http.StatusNotFound)
	return
}

c.Completed = *req.Completed
```

解析请求和校验是每次请求自己的工作，放在锁外。查找、修改和读取共享课程必须使用与其他接口相同的锁。

这里有两个不同的 nil 检查：req.Completed == nil 检查客户端是否提供了有效状态；c == nil 检查课程是否存在。前者返回 400，后者返回 404。

## 6 详细理解 c.Completed = *req.Completed

| 表达式 | 含义 | 类型 |
| --- | --- | --- |
| req | 本次请求结构体的值 | UpdateCourseCompletedRequest |
| req.Completed | 请求里的指针字段 | *bool |
| *req.Completed | 取出指针指向的布尔值 | bool |
| c | 指向切片中课程的指针 | *Course |
| c.Completed | 原课程的完成状态字段 | bool |

假设请求传来 true：

```text
req.Completed → 指向 true
       ↓ 解引用
*req.Completed 得到 true
       ↓ 赋值
c 指向的原课程，其 Completed 变成 true
```

右边的 *req.Completed 等价于 *(req.Completed)，先访问字段，再通过指针取值。

左边 c.Completed 等价于 (*c).Completed：访问结构体字段时，Go 允许自动解引用结构体指针。

可以拆成两句帮助理解：

```go
completed := *req.Completed
c.Completed = completed
```

不能直接写 c.Completed = req.Completed，因为左边是 bool，右边是 *bool。赋值复制的是 true 或 false，不是把请求中的指针保存到课程中。

## 7 构造响应

当前代码：

```go
resData := FindCourseResData{
	Code:    0,
	Data:    *c,
	Message: "修改成功",
}

data, err := json.Marshal(resData)
if err != nil {
	http.Error(w, "编码失败", http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_, err = w.Write(data)
if err != nil {
	log.Println("响应写入失败：", err)
}
```

Data 要求 Course 值，因此通过 *c 取得修改后的课程值。先完成编码，再提交 HTTP 200；Marshal 是编码，失败返回 500。Write 失败时仅记录日志，不追加第二份错误响应。

本接口当前 JSON 业务 code 为 0，HTTP 状态为 200，两者独立。其他成功接口当前业务 code 使用 200 或 201，因此暂时存在约定差异；本次保留原写法。前端联调前需要统一成功码约定，不能假设整个项目都使用同一个业务成功码。

## 8 curl 练习

修改代码后重启服务，将 ID 1 的课程设为完成：

```sh
curl -i -X PATCH http://localhost:8080/courses/1/completed \
  -H 'Content-Type: application/json' \
  -d '{"completed":true}'
```

取消完成：

```sh
curl -i -X PATCH http://localhost:8080/courses/1/completed \
  -H 'Content-Type: application/json' \
  -d '{"completed":false}'
```

两种请求都应该返回 200，data 中的 completed 分别为 true 和 false，ID 和名称不变。再次 GET /courses/1 可以核对真实存储的状态。

继续测试：

| 请求 | 预期 |
| --- | --- |
| {} | 400，不改变原状态 |
| {"completed":null} | 400，不改变原状态 |
| {"completed":"yes"} | 400，不改变原状态 |
| {"completed":0} | 400，不改变原状态 |
| 非法 JSON、空请求体 | 400，不改变原状态 |
| 不存在的 ID，加有效布尔状态 | 404 |

## 9 本次检查与范围

格式、go vet 和编译检查通过。go test 显示 [no test files]，当前没有持久化自动化测试文件。

用当前源码的临时副本，仅替换监听地址为临时本机端口，以 go build -race 构建并验证：

- true、false 能正确写入，并可通过详情接口读到。
- 额外传入 ID 和名称不会改动课程的对应字段。
- 缺失字段、null、错误类型、非法 JSON、数组及空请求体返回 400，原课程保持不变。
- 不存在及已删除的课程返回 404。
- 并发执行查询、名称修改和状态修改时，响应正常，竞态检测器未报告数据竞争或 panic。

这些是实际接口检查，不等于已有持久化单元测试，也不证明所有未来并发路径都安全。未进行前端联调，未打断原来运行的服务。

仍保留前几节的教学范围：数据只在内存中；未限制请求体大小或检查尾随 JSON；锁覆盖响应写入。当前 Decode 默认忽略未知字段，不是严格禁止额外字段的接口。
