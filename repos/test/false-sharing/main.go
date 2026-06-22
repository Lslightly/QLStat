// Copyright 2026 Qingwei Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
https://www.zhihu.com/question/1983394937543366516
*/
package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const N = math.MaxInt32

func calc(numCPU int) {
	runtime.GOMAXPROCS(numCPU)

	// 随机数种子
	seed := time.Now().UnixNano()

	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup

	// 创建通道收集每个goroutine的结果
	results := make(chan int, numCPU)

	// 计算每个goroutine需要处理的数量
	batchSize := N / numCPU

	for workerID := 0; workerID < numCPU; workerID++ {
		start := workerID * batchSize
		end := start + batchSize

		// 最后一个goroutine: 边界检查 (整数除法的截断)
		if workerID == numCPU-1 {
			end = N
		}

		// 为每个goroutine创建独立的随机数生成器
		localRand := rand.New(rand.NewSource(seed + int64(workerID)))
		// 每次迭代创建一个 localCount，被一个 goroutine 使用（安全）
		localCount := 0

		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := start; i < end; i++ {
				x := localRand.Float64() - 0.5
				y := localRand.Float64() - 0.5

				if x*x+y*y < 0.25 {
					localCount++
				}
			}
			results <- localCount
		}()
	}

	// 等待所有goroutine完成并关闭结果通道
	go func() {
		wg.Wait()
		close(results)
	}()

	// 汇总所有结果
	totalCount := 0
	for count := range results {
		totalCount += count
	}

	// 计算π的近似值
	pi := float64(totalCount) / float64(N) * 4.0
	fmt.Printf("π ≈ %.15f\n", pi)
}

func main() {
	calc(1)
}
