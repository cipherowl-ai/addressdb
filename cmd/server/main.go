package main

/**
Example server for the Bloom filter, should be good enough as OOB solution


*/
import (
	"addressdb/store"
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"strconv"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

var (
	filter    *bloom.BloomFilter
	logger    = log.New(os.Stdout, "BloomServer: ", log.LstdFlags)
	lasterror error
	ratelimit int
	burst     int
)

type Response struct {
	Query   string `json:"query"`
	InSet   bool   `json:"in_set"`
	Message string `json:"message"`
}

func main() {
	filename := flag.String("f", "bloomfilter.gob", "Path to the .gob file containing the Bloom filter")
	port := flag.Int("p", 8080, "Port to listen on")
	ratelimit_v := flag.Int("r", 20, "Ratelimit")
	burst_v := flag.Int("b", 5, "Burst")
	flag.Parse()

	// Use the values
	ratelimit = *ratelimit_v
	burst = *burst_v

	filter, lasterror = store.BloomFilterFromFile(*filename)

	if lasterror != nil {
		logger.Fatalf("Failed to load Bloom filter: %v", lasterror)
		os.Exit(-1)
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/check", rateLimitMiddleware(http.HandlerFunc(checkHandler))).Methods("GET")
	r.Handle("/checkBatch", rateLimitMiddleware(http.HandlerFunc(checkBatchHandler))).Methods("POST")

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(*port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	gracefulShutdown(srv)
}
func checkHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("s")
	if query == "" {
		http.Error(w, `{"error": "Missing 's' parameter"}`, http.StatusBadRequest)
		return
	}

	found := filter.TestString(query)
	response := struct {
		Found bool `json:"found"`
	}{
		Found: found,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func checkBatchHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	var requestBody struct {
		Addresses []string `json:"addresses"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, `{"error": "Invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Check if the addresses list is empty
	if len(requestBody.Addresses) == 0 {
		http.Error(w, `{"error": "Empty addresses list"}`, http.StatusBadRequest)
		return
	}

	// Check each address against the Bloom filter
	found := make([]string, 0)
	notFound := make([]string, 0)

	for _, address := range requestBody.Addresses {
		if filter.TestString(address) {
			found = append(found, address)
		} else {
			notFound = append(notFound, address)
		}
	}

	var resultsMerged struct {
		Found         []string `json:"found"`
		NotFound      []string `json:"notfound"`
		FoundCount    int      `json:"found_count"`
		NotFoundCount int      `json:"notfound_count"`
	}

	resultsMerged.Found = found
	resultsMerged.NotFound = notFound
	resultsMerged.FoundCount = len(found)
	resultsMerged.NotFoundCount = len(notFound)

	// Prepare the response
	response := resultsMerged
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(ratelimit), burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func gracefulShutdown(srv *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	logger.Println("shutting down")
	os.Exit(0)
}