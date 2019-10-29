package pools

import (
	"fmt"
	"testing"
)

func TestGoPool(t *testing.T) {
	testGoPoolNum(10, 500)
	testGoPoolNum(10, 50000)
}

type dataInfo struct {
	poolNum      int
	maxSendCnt   int
	curSendIndex int
}

func testGoPoolNum(poolNum, sendCnt int) {
	fmt.Printf("poolNum:%d, sendCnt:%d.\r\n", poolNum, sendCnt)
	gopool := NewGoPool(poolNum)
	defer gopool.Destroy()
	//此次特意将data作为一个相对于gopool来讲的全局变量。这样可以通过在打印中找到相同的两项来证明多个goroutine都在同时工作。
	data := &dataInfo{
		poolNum:      poolNum,
		maxSendCnt:   sendCnt,
		curSendIndex: 0,
	}
	for index := 0; index < sendCnt; index++ {
		data.curSendIndex++
		gopool.AddWorker(data, func(d interface{}) {
			di := d.(*dataInfo)
			fmt.Printf("poolNum:%d ,maxSendCnt:%d ,curSendIndex:%d.\r\n", di.poolNum, di.maxSendCnt, di.curSendIndex)
		})
	}
}
