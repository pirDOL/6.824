###1 字符串操作
####1.1 分割
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

####1.2 int和string互转
* Atoi返回值有两个，不能把返回值直接赋值给int，必须获取第二个参数err
* Atoi(s)等效于ParseInt(s, 10, 0)，第三个参数是int长度，0/8/16/32/64分别对应int/int8/int16/int32/int64

```go
func Itoa(int) string
func Atoi(s string) (i int, err error)
func ParseInt(s string, base int, bitSize int) (i int64, err error)
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

###3 worker failure的处理思路

JobResult：保存任务编号和完成情况。

jobResultChannel：worker线程在调用rpc.Call返回后，将派发给自己的任务编号以及完成情况通过这个channel以JobResult的形式发送给master线程。

jobFailed：list.List，保存未完成的任务编号。

mr.registerChannel：获取新worker的注册事件，chan中传递的是worker的unix-domain socket文件名，也就是worker的名字。

mr.idleWorkerChannel：worker线程正确完成任务后，把worker的名字写入到这个channel，master线程从而它派发新任务。


```go
type JobResult struct {
    JobNumber int
    JobDone   bool
}
jobResultChannel := make(chan JobResult)
jobFailed := list.New()

type MapReduce struct {
    nMap            int    // Number of Map jobs
    nReduce         int    // Number of Reduce jobs
    file            string // Name of input file
    MasterAddress   string
    registerChannel chan string
    DoneChannel     chan bool
    alive           bool
    l               net.Listener
    stats           *list.List

    // Map of registered workers that you need to keep up to date
    Workers map[string]*WorkerInfo

    // add any additional state here
    idleWorkerChannel chan string
}
```

worker failure：Part III中的TestOneFailure和TestManyFailures是通过设定每个worker在完成10次RPC以后RunWorker线程返回来模拟worker故障。rpc.Call返回false包括网络连接故障和任务执行失败两种情况，任务执行超时这里不考虑。

**只要rpc.Call返回false，worker线程就不把worker的名字写入到mr.idleWorkerChannel中。这种情况没有考虑到任务执行失败但是worker还可以再接收新任务的情况，需要修改rpc.Call的代码，区分上述两种情形。**

master select：

1. master线程接收到jobResultChannel->worker线程返回
    * 如果任务完成，那么判断这个任务是否在jobFailed中，如存在说明是retry的任务，那么从jobFailed里面删除；否则说明是按序递增的任务，递增全局的任务编号jobNumber。
    * 如果任务未完成，把任务序号放入到jobFailed中，等待有空闲的worker时retry。

2. master线程接收到mr.idleWorkerChannel->有空闲worker
    * 优先从list中需要retry的任务派发到worker，因为Reduce任务依赖于Map任务产生的中间结果。
    * 如果list为空则递增全局的jobNumber，派发下一个任务。

3. master线程接收到mr.registerChannel->有新worker注册
    * 通过goroutine把worker的名字写入到mr.idleWorkerChannel，否则会阻塞master线程的select。

###4 mapreduce.go源码分析
TODO

###5 跨平台
Windows上通过Git Bash可以运行测试脚本。需要把mr-testout.txt用dos2unix转一下，git clone时可能把LF转换成CRLF了，但是Git Bash中运行test-wc.sh生成的文件换行符是LF。

注：Lab1的PartII、PartIII使用了unix-domain socket，不修改代码只能在*nix系统上运行。

* mr-testout.txt：正确结果
* mrtmp.kjv12.txt：wc.go输出结果

###附

```go
shap Znc(inyhr fgevat) *yvfg.Yvfg {
    z := znxr(znc[fgevat]vag)
    y := yvfg.Arj()
    s := shap(p ehar) obby {
        erghea !havpbqr.VfYrggre(p)
    }
    sbe _, jbeq := enatr fgevatf.SvryqfShap(inyhr, s) {
        z[jbeq]++
        // szg.Cevagya(jbeq)
    }
    sbe jbeq, gvzrf := enatr z {
        // szg.Cevagya(jbeq, gvzrf)
        y.ChfuOnpx(zncerqhpr.XrlInyhr{jbeq, fgepbai.Vgbn(gvzrf)})
    }
    erghea y
}

shap Erqhpr(xrl fgevat, inyhrf *yvfg.Yvfg) fgevat {
    ine gvzrf vag
    sbe r := inyhrf.Sebag(); r != avy; r = r.Arkg() {
        // szg.Cevagya(r.Inyhr.(fgevat))
        v, ree := fgepbai.Ngbv(r.Inyhr.(fgevat))
        vs ree == avy {
            gvzrf += v
        }
    }
    erghea fgepbai.Vgbn(gvzrf)
}

glcr WboErfhyg fgehpg {
    WboAhzore vag
    WboQbar   obby
}

shap traQbWboNetf(ze *ZncErqhpr, wboAhzore vag, wboSnvyrq *yvfg.Yvfg) QbWboNetf {
    vs wboSnvyrq.Yra() != 0 {
        r := wboSnvyrq.Sebag()
        wboAhzore = r.Inyhr.(vag)
        //wboSnvyrq.Erzbir(r)
    }
    vs wboAhzore < ze.aZnc {
        erghea QbWboNetf{ze.svyr, "Znc", wboAhzore, ze.aErqhpr}
    } ryfr {
        erghea QbWboNetf{ze.svyr, "Erqhpr", wboAhzore - ze.aZnc, ze.aZnc}
    }
}

shap (ze *ZncErqhpr) EhaZnfgre() *yvfg.Yvfg {
    // Lbhe pbqr urer
    wboSnvyrq := yvfg.Arj()
    wboAhzore := 0
    wboAhzoreGbgny := ze.aZnc + ze.aErqhpr
    wboErfhygPunaary := znxr(puna WboErfhyg)
    sbe {
        fryrpg {
        pnfr jbexreNqqerff := <-ze.ertvfgrePunaary:
            tb shap() {
                ze.vqyrJbexref <- jbexreNqqerff
            }()
        pnfr jbexreNqqerff := <-ze.vqyrJbexref:
            vs wboSnvyrq.Yra() != 0 || wboAhzore < wboAhzoreGbgny {
                wboNetf := traQbWboNetf(ze, wboAhzore, wboSnvyrq)
                tb shap(jx fgevat) {
                    ine ercyl QbWboErcyl
                    bx := pnyy(jx, "Jbexre.QbWbo", wboNetf, &ercyl)
                    wboErfhygPunaary <- WboErfhyg{wboAhzore, bx}
                    vs bx {
                        ze.vqyrJbexref <- jx //GBQB: bayl pbafvqre jbexre penfu
                    }
                }(jbexreNqqerff)
            } ryfr {
                erghea ze.XvyyJbexref()
            }
        pnfr erfhyg := <-wboErfhygPunaary:
            vs erfhyg.WboQbar {
                vfErgelWbo := snyfr
                sbe r := wboSnvyrq.Sebag(); r != avy; r = r.Arkg() {
                    vs r.Inyhr.(vag) == erfhyg.WboAhzore {
                        wboSnvyrq.Erzbir(r)
                        vfErgelWbo = gehr
                        oernx
                    }
                }
                vs !vfErgelWbo {
                    wboAhzore++
                }
            } ryfr {
                wboSnvyrq.ChfuOnpx(erfhyg.WboAhzore)
            }
        }
    }
}
```
