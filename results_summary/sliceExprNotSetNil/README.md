# sliceNotSetNil结果

## commit #14863849

这里是commit #14863849的结果。文件说明如下。

```
[0] .
├── [1] README.md
├── [2] concat.csv # 原始数据
└── [3] concat_manual_annotated.csv # 人工确认检测到的代码是否符合sliceNotSetNil模式(WEF-sliceExpr)，添加一列manual-check，如果为true，则表示人工确认代码模式确实是有可能存在问题的。如果为false，则检测存在问题，有一些实际上会
```

在[concat_manual_annotated.csv](concat_manual_annotated.csv)添加两列为，一列为manualCheck，一列为improveChance。
- manualCheck为bool类型。如果为false，则表示报告是误报，如果为true，则为正确的报告。
- improveChange为bool类型。只有当manualCheck是false，即报告是误报的时候才可能为true，表示静态分析工具有可提升的可能。

在人工检查的时候的查看标准为:
1. 被切对象(`s[start:]`中0~start-1部分的对象)的引用是否在当前函数外就消失了，如果消失，则报告正确，如果没有消失，则报告错误
   1. 如果有其他变量引用被切对象(如果是获得被切对象的域值，不算引用)，则当变量可能逃出函数作用域时报告错误，否则按照没有其他变量引用被切对象处理
   2. 如果没有其他变量引用被切对象，则
      1. 如果切片表达式之后的控制流可能到达append函数，则报告错误
         1. 如果append添加的元素引用被切对象，那set不set nil无所谓，报告错误。如果没有引用，那也多报了，报告错误。
      2. 如果没有append，则报告正确

### 可能可以提升的点

/data/github_go/repos/EasyDarwin/rtsp/pusher.go:266 被局部变量引用，但是可能逃出函数。

```go
pack = pusher.queue[0]
pusher.queue = pusher.queue[1:]
...
pusher.BroadcastRTP(pack)
```

/data/github_go/repos/beats/libbeat/publisher/queue/diskqueue/queue.go:190 被append到其他对象中

```go
for len(initialSegments) > 0 && initialSegments[0].id < readSegmentID {
	ackedSegments = append(ackedSegments, initialSegments[0])
	initialSegments = initialSegments[1:]
}
```

/data/github_go/repos/benthos/internal/impl/aws/input_kinesis.go:533 不同的局部变量assign方式`range`

```go
var r *kinesis.Record
for i, r = range pending {
	if recordBatcher.AddRecord(r) {
		if pendingMsg, err = recordBatcher.FlushMessage(commitCtx); err != nil {
			k.log.Errorf("Failed to dispatch message due to checkpoint error: %v\n", err)
		}
		break
	}
}
if pending = pending[i+1:]; len(pending) == 0 {
	unblockPullChan()
}
```

/data/github_go/repos/benthos/internal/impl/aws/input_sqs.go:309 assign方式`send`

```go
case a.messagesChan <- pendingMsgs[0]:
	pendingMsgs = pendingMsgs[1:]
```

/data/github_go/repos/benthos/internal/impl/aws/input_s3.go:397 局部变量被return

```go
obj := s.pending[0]
s.pending = s.pending[1:]
return obj, nil
```

/data/github_go/repos/aws-nuke/resources/elb-elb.go:62 `elbNames[:requestElements]`被作为参数传入函数。

```go
requestElements := len(elbNames)
if requestElements > 20 {
	requestElements = 20
}

tagResp, err := svc.DescribeTags(&elb.DescribeTagsInput{
	LoadBalancerNames: elbNames[:requestElements],
})
if err != nil {
	return nil, err
}
for _, elbTagInfo := range tagResp.TagDescriptions {
	elb := elbNameToRsc[*elbTagInfo.LoadBalancerName]
	resources = append(resources, &ELBLoadBalancer{
		svc:  svc,
		elb:  elb,
		tags: elbTagInfo.Tags,
	})
}

// Remove the elements that were queried
elbNames = elbNames[requestElements:]
```

/data/github_go/repos/benthos/internal/impl/aws/input_s3.go:635

```go
for len(dudMessageHandles) > 0 {
	input := sqs.ChangeMessageVisibilityBatchInput{
		QueueUrl: aws.String(s.conf.SQS.URL),
		Entries:  dudMessageHandles,
	}

	// trim input entries to max size
	if len(dudMessageHandles) > 10 {
		input.Entries, dudMessageHandles = dudMessageHandles[:10], dudMessageHandles[10:]
	} else {
		dudMessageHandles = nil
	}
	_, _ = s.sqs.ChangeMessageVisibilityBatch(&input)
}
```


/data/github_go/repos/LeetCode-Go/leetcode/0102.Binary-Tree-Level-Order-Traversal/102. Binary Tree Level Order Traversal.go:38 被切对象还在其他对象中引用。

```go
l := len(queue)
tmp := make([]int, 0, l)
for i := 0; i < l; i++ {
	if queue[i].Left != nil {
		queue = append(queue, queue[i].Left)
	}
	if queue[i].Right != nil {
		queue = append(queue, queue[i].Right)
	}
	tmp = append(tmp, queue[i].Val)
}
queue = queue[l:]
res = append(res, tmp)
```

> /data/github_go/repos/LeetCode-Go/leetcode有较多上述这种前面拿出来放到后面的情况

/data/github_go/repos/LeetCode-Go/leetcode/0559.Maximum-Depth-of-N-ary-Tree/559.Maximum Depth of N-ary Tree.go:24 其中`ele`实际上并没有整体被其他对象引用，只有`Children`slice value被copy了，这单个结点是可以释放的，积少成多。不过q已经被append了。

```go
for length != 0 {
	ele := q[0]
	q = q[1:]
	length--
	if ele != nil && len(ele.Children) != 0 {
		q = append(q, ele.Children...)
	}
}
```

/data/github_go/repos/VictoriaMetrics/lib/logstorage/storage.go:324 在for循环里面set nil，不属于同一个block，这种情况是好情况。

```go
for i := range ptwsToDelete {
	s.partitions[i] = nil
}
s.partitions = s.partitions[len(ptwsToDelete):]
```

/data/github_go/repos/VictoriaMetrics/lib/logstorage/storage_search.go:157 ptws实际上是其他slice的切片，只是用来做统计操作，set nil没有意义。

```go
ptws := s.partitions
minDay := tf.minTimestamp / nsecPerDay
n := sort.Search(len(ptws), func(i int) bool {
	return ptws[i].day >= minDay
})
ptws = ptws[n:]
maxDay := tf.maxTimestamp / nsecPerDay
n = sort.Search(len(ptws), func(i int) bool {
	return ptws[i].day > maxDay
})
ptws = ptws[:n]
```

/data/github_go/repos/VictoriaMetrics/lib/mergeset/table.go:563 pws实际上函数传递的参数，调用者可能还有对被切对象的引用。

```go
func (tb *Table) mergePartsOptimal(pws []*partWrapper) error {
	sortPartsForOptimalMerge(pws)
	for len(pws) > 0 {
		n := defaultPartsToMerge
		if n > len(pws) {
			n = len(pws)
		}
		pwsChunk := pws[:n]
		pws = pws[n:]
		err := tb.mergeParts(pwsChunk, nil, true)
		if err == nil {
			continue
		}
		tb.releasePartsToMerge(pws)
		return fmt.Errorf("cannot optimally merge %d parts: %w", n, err)
	}
	return nil
}
```

> 综上，如果有变量引用，那么如果这个变量没有逃逸，就可以进行下一步判断append操作，如果没有append操作，实际上就可以添加nil以加速回收。

/data/github_go/repos/awesome-golang-algorithm/leetcode/1-100/0100.Same-Tree/Solution.go:31 slice的生命期太短了

```go
func isSameTree2(p *TreeNode, q *TreeNode) bool {
	qP := []*TreeNode{p}
	qQ := []*TreeNode{q}

	for len(qP) != 0 && len(qQ) != 0 {
		pNode := qP[0]
		qP = qP[1:]

		qNode := qQ[0]
		qQ = qQ[1:]

		if pNode == nil && qNode == nil {
			continue
		}
		if pNode == nil && qNode != nil || pNode != nil && qNode == nil {
			return false
		}
		if pNode.Val != qNode.Val {
			return false
		}

		qP = append(qP, pNode.Left, pNode.Right)
		qQ = append(qQ, qNode.Left, qNode.Right)
	}

	if len(qP) == 0 && len(qQ) == 0 {
		return true
	}
	return false
}
```

/data/github_go/repos/aws-nuke/resources/s3-buckets.go:191 slice没有引用但是被append。

```go
func (iter *s3DeleteVersionListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
	}

	if len(iter.objects) == 0 && iter.Paginator.Next() {
		output := iter.Paginator.Page().(*s3.ListObjectVersionsOutput)
		iter.objects = output.Versions

		for _, entry := range output.DeleteMarkers {
			iter.objects = append(iter.objects, &s3.ObjectVersion{
				Key:       entry.Key,
				VersionId: entry.VersionId,
			})
		}
	}

	return len(iter.objects) > 0
}
```

/data/github_go/repos/beego/server/web/server.go:393 slice值判断存在问题

```go
entryPointTree.fixrouters[i].leaves[0] = nil
entryPointTree.fixrouters[i].leaves = entryPointTree.fixrouters[i].leaves[1:]
```
