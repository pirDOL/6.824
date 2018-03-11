### PartI
#### 排序
reduce从mrtmp.${job_name}-{0..nMap}-iReduce共nMap个文件按照key排序，再执行ReduceF

#### mrtmp文件如何组织
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

### PartII&V
1. 文件中存在隐藏字符导致代码diff比较失败、bash脚本执行异常

repo是clone到在windows，通过virtualbox的共享文件夹在ubuntu中调试，mr-test.txt、mr-challenge.txt、test-xxx.sh中\n被替换成\r\n。
