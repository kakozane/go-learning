# 第十一节 DELETE 删除课程、切片赋值与 204 响应

上一篇：[第十节 PATCH 修改课程名称与指针响应](./10-patch-course-name-and-pointer-response.md)。

本节完成 DELETE /courses/{id}。删除成功返回 204，不存在返回 404，课程的 HTTP 增删改查流程至此齐备。数据仍然只保存在内存中。

## 1 删除接口的执行顺序

```go
mux.HandleFunc("DELETE /courses/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id错误或不能为空", http.StatusBadRequest)
		return
	}

	coursesMu.Lock()
	defer coursesMu.Unlock()

	remaining, deleted := deleteCourseByID(courses, id)
	if !deleted {
		http.Error(w, "课程不存在", http.StatusNotFound)
		return
	}

	courses = remaining
	w.WriteHeader(http.StatusNoContent)
})
```

路径中的 ID 已经指定要删除谁，不需要读取 JSON 请求体。加锁保护查找、删除及切片赋值，函数退出前释放锁。

当前带 {id} 的路由要求存在对应路径段，空 ID 检查是额外防御，不代表 DELETE /courses 会进入这个处理函数。

## 2 复用之前的删除函数

```go
func deleteCourseByID(courses []Course, id string) ([]Course, bool) {
	for i, course := range courses {
		if course.ID == id {
			newCourses := append(courses[:i], courses[i+1:]...)
			return newCourses, true
		}
	}
	return courses, false
}
```

两个返回值分别是删除后的切片，以及是否找到了并删除课程。没找到时返回原切片和 false。

本接口把 false 解释为课程不存在，因此用 404，而不是含糊的“删除失败，请重试”或参数错误 400。

## 3 为什么不能忽略第一个返回值

之前的错误写法：

```go
_, deleted := deleteCourseByID(courses, id)
```

函数修改了共享的底层数组，却没有更新外层 courses 变量的长度。需要接住新切片，再赋值：

```go
remaining, deleted := deleteCourseByID(courses, id)
if !deleted {
	// 返回 404 并退出
	return
}
courses = remaining
```

和 courses = append(courses, course) 一样，返回值携带了更新后的切片描述信息。

## 4 [1, 3, 3] 是怎么来的

这里用课程 ID 简写，实际数组元素是 Course 结构体。

```text
原数组：[1, 2, 3]
下标：   0  1  2
```

删除 ID 2 时，i 等于 1：

```go
courses[:i]   // [1]
courses[i+1:] // [3]
```

append 将后面的课程 3 追加到前面 [1] 的末尾。前面的切片仍有足够容量，因此复用原底层数组，把下标 1 的课程 2 覆盖成课程 3：

```text
底层数组：[1, 3, 3]
新切片： [1, 3]       长度 2
旧切片： [1, 3, 3]    长度仍是 3
```

原数组最后一个位置没有被这句 append 清空。删除操作不是消灭数组中的一个格子，而是移动元素，并用更短的切片表示有效范围。

执行 courses = remaining 后，遍历和 JSON 编码只会处理长度范围内的 [1, 3]。底层数组的容量不必因此缩小。

当前删除实现保留了底层数组尾部的旧值；本节关注可见范围和长度，不把它理解成主动清空底层存储。

## 5 是因为没使用指针才出问题吗

不是函数设计错误，而是调用方没有接住返回值。

Go 传递切片时复制描述信息，包括长度、容量和底层数组的位置。通过副本能修改共享数组中的元素，但不能直接替换调用方的切片变量。

有两种合法设计：

| 对比 | 返回切片 | 传入切片指针 |
| --- | --- | --- |
| 参数 | []Course | *[]Course |
| 谁更新外层变量 | 调用方接住返回值后赋值 | 函数内部通过指针赋值 |
| 调用形式 | deleteCourseByID(courses, id) | deleteCourseByID(&courses, id) |
| 注意事项 | 别忽略返回的切片 | 不要传 nil 指针，理解内部修改行为 |

两种写法都可以修改底层数组，都需要保护共享数据，也不能简单断言指针版性能更好。当前项目继续使用返回值方式。

## 6 指针版作为对照练习

下面是替代示例，不是项目当前实现。不能在同一包内同时保留两个同名普通函数。

```go
func deleteCourseByID(courses *[]Course, id string) bool {
	for i, course := range *courses {
		if course.ID == id {
			*courses = append((*courses)[:i], (*courses)[i+1:]...)
			return true
		}
	}
	return false
}
```

- courses 的类型是 *[]Course，指向一个切片变量。
- *courses 取出它指向的切片。
- (*courses)[:i] 先取出切片，再截取元素。
- *courses = ... 更新调用方的切片变量，包括长度。

调用时在锁内传入 &courses：

```go
deleted := deleteCourseByID(&courses, id)
```

这时不再需要 courses = remaining。指针改变的是谁负责更新切片变量，不会让底层数组自动缩小。示例要求传入 &courses 这样的非 nil 指针。

## 7 没有响应体，前端怎么知道成功

HTTP 响应除了响应体，还有状态码。204 No Content 本身就表达“请求成功，没有响应体”。

```go
w.WriteHeader(http.StatusNoContent)
```

后面不需要 Marshal 或 Write，也不需要设置 JSON Content-Type。本次检查移除了成功删除响应中多余的 JSON 响应头。

前端可以根据状态码处理。以下是 React 示例，尚未接入当前项目：

```javascript
const response = await fetch(`/courses/${id}`, {
  method: 'DELETE',
})

if (response.status === 204) {
  setCourses(prev => prev.filter(course => course.id !== id))
} else {
  const message = await response.text()
  console.error('删除失败：', message)
}
```

不要对 204 响应调用 response.json()，因为没有 JSON 内容可解析。上述相对地址假定请求能通过同源部署或代理访问 Go 服务，实际前后端联调时还需要配置地址。

也可以约定成功返回 200 加 JSON 消息，但不能一边返回 204，一边又写成功 JSON。当前选择 204。

## 8 curl 验证

重启后端，在另一个终端执行：

```sh
curl -i -X DELETE http://localhost:8080/courses/2
```

预期状态为 204，响应体为空。然后查询：

```sh
curl -i http://localhost:8080/courses
curl -i http://localhost:8080/courses/2
```

列表应该只有 ID 1 和 3，详情返回 404。

重复删除：

```sh
curl -i -X DELETE http://localhost:8080/courses/2
```

按照当前接口约定，已经不存在，因此返回 404。继续删除剩余课程，列表的 data 应为 []。再新增课程应正常工作，nextID 不会因为删除而倒退。

## 9 本次验证与范围

格式检查、go vet、编译通过。go test 显示 [no test files]，当前没有持久化的自动化测试文件。

使用当前源码临时副本，仅替换为临时本机监听端口，并通过 go build -race 构建，验证了：

- 删除中间、最后、唯一一门课程，列表长度和剩余 ID 正确。
- 删除成功返回 204 和空响应体。
- 重复删除、不存在的 ID、空列表删除返回 404，不改变列表。
- 已删除课程的详情返回 404。
- 删除全部课程后列表为 []，仍可继续新增。
- 交错并发执行 GET、POST、DELETE，针对同一 ID 仅一次删除成功；最终数量正确，ID 不重复。
- 本次请求期间没有竞态或 panic 报告。

未打断原有服务，未进行前端联调或生产压测。课程仍仅存于内存，重启会恢复初始课程；现有锁覆盖响应写入等教学限制仍保留。
