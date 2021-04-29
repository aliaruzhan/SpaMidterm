package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)
const Th = 6
// сюда писать код
func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, i := range jobs{
		wg.Add(1)
		out := make(chan interface{})
		go func(in chan interface{}, out chan interface{}, wg *sync.WaitGroup, i job) {
			defer wg.Done()
			defer close(out)
			i(in, out)
		}(in, out, wg, i)

		in = out
	}
	wg.Wait()
}

//SingleHash считает значение crc32(data)+"~"+crc32(md5(data))
func SingleHash(in chan interface{}, out chan interface{}) {
	swg := &sync.WaitGroup{}
	mx := &sync.Mutex{}

	for data := range in {
		swg.Add(1)
		go SingleHashCalc(data, out, swg, mx)
	}
	swg.Wait()
}

func SingleHashCalc(data interface{}, out chan interface{},swg *sync.WaitGroup, mx *sync.Mutex) {
	defer swg.Done()
	convData := strconv.Itoa(data.(int))

	mx.Lock()
	md5data := DataSignerMd5(convData)
	mx.Unlock()

	crc32Chan := make(chan string)
	go func(data string, out chan string) {
		defer close(out)
		out <- DataSignerCrc32(data)
	}(convData, crc32Chan)

	md5crc32Data := DataSignerCrc32(md5data)
	crc32Data := <- crc32Chan

	out <- crc32Data + "~" + md5crc32Data
}

//MultiHash считает значение crc32(th+data))
func MultiHash(in chan interface{}, out chan interface{}) {
	mwg := &sync.WaitGroup{}

	for data := range in{
		mwg.Add(1)
		go MultiHashCalc(data, out, mwg)
	}
	mwg.Wait()
}

func MultiHashCalc(data interface{}, out chan interface{}, mwg *sync.WaitGroup){
	defer mwg.Done()

	wgr := &sync.WaitGroup{}
	res := make([]string, Th)
	passData := data.(string)

	for i := 0; i < Th; i++ {
		wgr.Add(1)
		convData := strconv.Itoa(i) + passData

		go func(data string, wg *sync.WaitGroup, res []string, i int){
			defer wgr.Done()
			res[i] = DataSignerCrc32(data)
		}(convData, wgr, res, i)
	}
	wgr.Wait()

	out <- strings.Join(res, "")
}

func CombineResults(in chan interface{}, out chan interface{}) {
	var arr []string

	for data := range in{
		arr = append(arr, data.(string))
	}
	sort.Strings(arr)

	out <- strings.Join(arr,"_")
}
