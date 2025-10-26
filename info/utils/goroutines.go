package utils

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type GoroutineInfo struct {
	start time.Time
	end   time.Time
}

// Все горутины в конкуренции
func RunAllExamples() {
	var wg sync.WaitGroup

	// map для логирования старта и конца goroutines
	infoMap := make(map[string]GoroutineInfo)
	var mu sync.Mutex

	// список функций с именами
	examples := map[string]func(){
		"simpleGoroutine": simpleGoroutine,
		"withWaitGroup":   withWaitGroup,
		"channelParam":    channelParam,
		"returnChannel":   returnChannel,
		"withRange":       withRange,
		"withSelect":      withSelect,
		"withErrGroup":    withErrGroup,
		"mergeChannels":   mergeChannels,
		"runTasks":        runTasks,
	}

	wg.Add(len(examples))
	// Работает не стабильно надо подумать как выводить анализ без пересечения горутин
	for name, ex := range examples {
		go func(n string, f func()) {
			defer wg.Done()
			// фиксируем старт
			start := time.Now()
			f()
			// фиксируем конец
			end := time.Now()

			mu.Lock()
			infoMap[n] = GoroutineInfo{start: start, end: end}
			mu.Unlock()
		}(name, ex)
	}

	wg.Wait()

	fmt.Println("\nВсе примеры завершены. Время старта и завершения goroutines:")
	for name, info := range infoMap {
		fmt.Printf("%s -> start: %s, end: %s, duration: %s\n",
			name,
			info.start.Format("15:04:05.000"),
			info.end.Format("15:04:05.000"),
			info.end.Sub(info.start),
		)
	}
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

// Канал как параметр функции
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

// Возврат канала из функции (fan-out pattern)
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

// Range + закрытие канала
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

// Select + timeout
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

// ErrGroup — параллельные HTTP-запросы
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

	if useResp != nil {
		useResp.Body.Close()
	}
	if taskResp != nil {
		taskResp.Body.Close()
	}

	fmt.Println("User:", useResp.Status)
	fmt.Println("Task:", taskResp.Status)

}

// Передача одного канала в другую горутину
// Что то из базового понимания пайплайнов, но это более сложная тема
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

	// Читаем ВСЕ результаты до завершения
	merged := fanIn(a, b)
	for res := range merged {
		fmt.Println("Merged:", res)
	}
}

func runTasks() {
	var wg sync.WaitGroup
	ch := make(chan int)

	// Большая goroutine (телега)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sum := 0
		for i := 0; i < 1_000_000; i++ {
			sum += i
			if i%100_000 == 0 {
				ch <- sum
			}
		}
		close(ch)
	}()

	// Обработчик канала (читаем чанки)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for val := range ch {
			fmt.Println("Чанк результата:", val)
		}
	}()

	// Маленькие goroutines (шоколадки)
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Println("Шоколадка", id, "готова")
		}(i)
	}

	// Ждём завершения всех goroutines
	wg.Wait()
}
