package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/sonohgong/gosh/pkg/gosh2"
)

func handler(data interface{}) error {
	time.Sleep(time.Duration(rand.Int63n(100) * int64(time.Millisecond)))
	fmt.Println("Doing some work", time.Now())
	return nil
}

func main() {
	taskManager := gosh2.NewTaskManager(handler)
	taskManager.NewTasks(100000)

	scheduler := gosh2.NewScheduler(10*time.Millisecond, taskManager.StartTasks)
	scheduler.Start()

	taskManager.Wait()
	fmt.Println("task manager done!")

	scheduler.Stop()
	fmt.Println("done!")

	f, err := os.Create("memprof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
