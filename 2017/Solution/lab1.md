### PartI
#### reduce对中间结果按key排序
mapreduce论文3.1节reduce需要对mrtmp.${job_name}-{0..nMap}-iReduce共nMap个文件按照key排序，这是考虑同一个map任务中会产生相同的key（例如单词计数中一个map任务中一个单词出现了多次）并且不同的map任务也会产生相同的key，reduce需要把key相同的value（不论这些kv是来自同一个map任务还是多个map任务）merge在一起调用ReduceF。

实际上，按照key做unique不是一定要通过对key排序实现，如果kv比较小可以用一个map<key, []value>来实现。

>5. When a reduce worker is notified by the master
about these locations, it uses remote procedure calls
to read the buffered data from the local disks of the
map workers. When a reduce worker has read all intermediate
data, it sorts it by the intermediate keys
so that all occurrences of the same key are grouped
together. The sorting is needed because typically
many different keys map to the same reduce task. If
the amount of intermediate data is too large to fit in
memory, an external sort is used.

#### mrtmp文件如何组织
文件名：mrtmp.${job_name}-${map_task_number}-${reduce_task_number}
reduce_task_number = key % ${reduce_task_number}

格式：这份文件在实验中可以是任意格式，只要doMap和doReduce保持一致即可，doMap写，doReduce读。

注：mrtmp-${job_name}-res-${reduce_task_number}这个文件格式必须和`func (mr *Master) merge()`保持一致：用`json.Decoder`把每个key的`ReduceF(key, []value)`的结果序列化为json写入文件。除了reduce任务只有处理一个key以外，这个文件本身不是一个json格式。

```
work@work-VirtualBox:~/6.824/2017/6.824-golabs-2017-dev/src/mapreduce$ go test -v -run TestSequential
=== RUN   TestSequentialSingle
master: Starting Map/Reduce task test
Merge: read mrtmp.test-res-0
master: Map/Reduce task completed
--- PASS: TestSequentialSingle (25.82s)
=== RUN   TestSequentialMany
master: Starting Map/Reduce task test
Merge: read mrtmp.test-res-0
Merge: read mrtmp.test-res-1
Merge: read mrtmp.test-res-2
master: Map/Reduce task completed
--- PASS: TestSequentialMany (25.44s)
PASS
ok      mapreduce       51.293s
work@work-VirtualBox:~/6.824/2017/6.824-golabs-2017-dev/src/mapreduce$ go test -v -run TestSequential
=== RUN   TestSequentialSingle
master: Starting Map/Reduce task test
Merge: read mrtmp.test-res-0
master: Map/Reduce task completed
--- PASS: TestSequentialSingle (13.60s)
=== RUN   TestSequentialMany
master: Starting Map/Reduce task test
Merge: read mrtmp.test-res-0
Merge: read mrtmp.test-res-1
Merge: read mrtmp.test-res-2
master: Map/Reduce task completed
--- PASS: TestSequentialMany (14.06s)
PASS
ok      mapreduce       27.694s
```

#### PartIII&IV
思路：

1. 调度器：事件循环，单线程避免对迭代器变量加锁
    1. 新worker注册：
        3. 退出点（所有任务都完成了）
        2. 调用迭代器的next接口获取下一个任务给worker执行
        3. 迭代器没有返回可以执行的任务，此时应该暂存这个worker
    2. rpc调用返回：每个rpc是一个goroutine，rpc的结果通过channel返回给事件循环
        2. 退出点
        2. rpc返回结果回调，失败的任务要存起来等待下次next时重试
        1. worker要启动一个goroutine写回注册channel，触发新任务执行
2. 迭代器：分发任务
    1. next：优先执行重试的任务，其次是顺序分发任务
    2. eof：没有正在执行的任务 && 所有任务都已经分发 && 没有等待重试的任务
    3. done：rpc返回结果的回调函数

TODO：worker的错误计数、标记不可用、周期探活。不可能worker故障一次就彻底放弃这个worker，应该通过连续错误计数超阈值标记为不可用，再周期探活。需要scheduler能保存全部的worker列表，而不是通过一个channel来回传递worker。

### PartII&V
问题：文件中存在隐藏字符导致代码diff比较失败、bash脚本执行异常。

原因：repo是clone到在windows，通过virtualbox的共享文件夹在ubuntu中调试，mr-test.txt、mr-challenge.txt、test-xxx.sh中\n被替换成\r\n。
