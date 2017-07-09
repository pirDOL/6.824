### 1 字符串分割
* strings.Fields：根据空格分割单词。
* strings.FieldsFunc：可以定制分割判断函数。strings.Fields等效于使用unicode.IsSpace划分。

```go
func Fields(s string) []string
func FieldsFunc(s string, f func(rune) bool) []string

f := func(c rune) bool {
    return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}
fmt.Printf("Fields are: %q", strings.FieldsFunc("  foo1;bar2,baz3...", f))
```

###2 list使用
####2.1 迭代时获取元素
e.Value.(type)：list为泛型容器，所以迭代时获取的Value是个接口，需要使用type assertion语法赋值到对应类型的变量中。

```go
l := list.New()
l.PushBack("123")
for e := l.Front(); e != nil; e = e.Next() {
    word := e.Value.(string)
}
```

####2.2 删除元素
Remove方法参数是迭代器，如果想从list中删除值为target的元素，必须遍历一遍。list的实现中迭代器失效的情况和C++中的vector一样，因此如果想删除后继续遍历需要在删除前保存当前节点的next。
具体源码分析参见[Go的List操作上的一个小“坑”](http://www.cnblogs.com/ziyouchutuwenwu/p/3780800.html)

```go
l := list.New()
l.PushBack(0)
l.PushBack(1)
l.PushBack(2)
l.PushBack(1)
l.PushBack(3)
target := 1
for e := l.Front(); e != nil; {
    if e.Value.(int) == a {
        d := e
        e = e.Next()
        l.Remove(d)
    } else {
        e = e.Next()
    }
}
```