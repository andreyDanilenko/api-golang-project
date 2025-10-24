package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Пример работы с range
func main() {
	log.Printf("Простая горутина")
	simpleGoroutine()

	log.Printf("Горутина с пакетом sync.WaitGroup")
	withWaitGroup()

	log.Printf("Канал как параметр")
	channelParam()

	log.Printf("Возвращаем канал из функции")
	returnChannel()

	log.Printf("Работаем с горутиной и range")
	withRange()

	log.Printf("Работаем с горутиной и select")
	withSelect()

	log.Printf("Работаем с горутиной и errorgroup")
	withErrGroup()

	log.Printf("Работаем с горутиной и errorgroup")
	mergeChannels()
}

// Простая горутина
func simpleGoroutine() {
	go fmt.Println("Привет из горутины!")
	time.Sleep(50 * time.Millisecond)
}

// С пакетом WaitGroup
func withWaitGroup() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("Задача 1 withWaitGroup done")
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Задача 2 withWaitGroup done")
	}()

	wg.Wait()
	fmt.Println("Все withWaitGroup done")
}

// ==========================================
// Канал как параметр функции
// ==========================================
func sendToChannel(ch chan<- string, msg string) {
	ch <- msg
}

// <-chan(только чтение)
// chan<-(только отправка)
func receiveFromChannel(ch <-chan string) {
	val := <-ch
	fmt.Println("Получил:", val)
}

func channelParam() {
	ch := make(chan string)
	go sendToChannel(ch, "Привет канал в параметре!")
	receiveFromChannel(ch)
}

// ==========================================
// Возврат канала из функции (fan-out pattern)
// ==========================================
func asyncFetch(url string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		resp, err := http.Get(url)
		if err != nil {
			ch <- fmt.Sprintf("error: %v", err)
			return
		}

		ch <- fmt.Sprintf("%s > %s", url, resp.Status)

	}()

	return ch
}

func returnChannel() {
	a := asyncFetch("https://httpbin.org/get")
	b := asyncFetch("https://httpbin.org/uuid")

	fmt.Println("Got a:", <-a)
	fmt.Println("Got b:", <-b)
}

// ==========================================
// Range + закрытие канала
// ==========================================
func withRange() {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for i := 1; i <= 3; i++ {
			ch <- fmt.Sprintf("Задача withRange #%d", i)
		}
	}()

	for val := range ch {
		fmt.Println("Got", val)
	}
}

// ==========================================
// Select + timeout
// ==========================================
func withSelect() {
	ch := make(chan string)
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- "done"
	}()

	select {
	case msg := <-ch:
		fmt.Println("Got: withSelect", msg)
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Timeout!")
	}
}

// ==========================================
// ErrGroup — параллельные HTTP-запросы
// ==========================================
func withErrGroup() {
	var g errgroup.Group
	var useResp, taskResp *http.Response

	g.Go(func() error {
		resp, err := http.Get("https://httpbin.org/get")
		useResp = resp
		return err
	})

	g.Go(func() error {
		resp, err := http.Get("https://httpbin.org/uuid")
		taskResp = resp
		return err
	})

	if err := g.Wait(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("User:", useResp.Status)
	fmt.Println("Task:", taskResp.Status)

}

// ==========================================
// Передача одного канала в другую горутину
// ==========================================
func fanIn(ch1, ch2 <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for ch1 != nil || ch2 != nil {
			select {
			case v, ok := <-ch1:
				if !ok {
					ch1 = nil
					continue
				}
				out <- v
			case v, ok := <-ch2:
				if !ok {
					ch2 = nil
					continue
				}
				out <- v
			}
		}
	}()
	return out
}

func mergeChannels() {
	a := asyncFetch("https://httpbin.org/get")
	b := asyncFetch("https://httpbin.org/uuid")
	for res := range fanIn(a, b) {
		fmt.Println("Merged:", res)
	}
}
