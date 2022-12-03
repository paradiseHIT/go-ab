package main

import "sync"

func main() {
	var wg sync.WaitGroup
	thread_num := 10
	wg.Add(thread_num)
	ab(thread_num, &wg)
	wg.Wait()
}
