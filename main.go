package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func TestEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Print("Handler started...")
	// emulate execution time, helpfull to confirm that handler finishes its work
	time.Sleep(8 * time.Second)
	w.WriteHeader(200)
	fmt.Fprint(w, "Test is what we usually do")
	log.Print("Handler finished")
}

func BackgroundWork(ctx context.Context) {
	i := 0
	for {
		i++
		log.Printf("Background Work: %d", i)
		time.Sleep(1 * time.Second)
		err := ctx.Err()
		if err != nil {
			log.Printf("Received cancelation message: %v", err)
			time.Sleep(1 * time.Second)
			log.Print("Background Work Cleaned up")
			// release goroutine
			return
		}
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/test", TestEndpoint).Methods("GET")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	done := make(chan os.Signal, 1)

	// Ctrl+C sends os.Interrupt
	// Docker and ECS sends syscall.SIGTERM
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// server closed by Shutdown()
			if err == http.ErrServerClosed {
				log.Print("Server Closed - Shutdown or Close called")
			} else {
				// timeout in Shutdown Context
				log.Fatalf("listen: %s\n", err)
			}
		}
	}()
	log.Print("Server Started")

	// context for background work
	bkgCtx := context.Background()
	bkgCtx, bkgCancel := context.WithCancel(bkgCtx)
	// background work
	go BackgroundWork(bkgCtx)

	// execution blocked here until application 
	sig := <-done
	log.Printf("Received Signal: %v, starting shutdown", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		log.Print("extra handling")

		// extra handling here
		bkgCancel()
		// todo: fix unnecessary wait
		time.Sleep(2 * time.Second)

		// Even though ctx will be expired, it is good practice to call its
		// cancellation function in any case. Failure to do so may keep the
		// context and its parent alive longer than necessary.
		cancel()
		log.Print("everything is clean")
	}()

	log.Print("Shutting Down Server ...")
	// blocked until all active requests finish or timeout
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
