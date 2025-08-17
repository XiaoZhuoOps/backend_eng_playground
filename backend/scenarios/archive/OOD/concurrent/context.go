package OOD

import (
	"fmt"
	"net/http"
	"time"
)

// result struct holds the response and any error encountered during the external request.
type result struct {
	resp *http.Response
	err  error
}

// hello handles HTTP requests by making a network request to an external API.
// It waits for the network request to finish but returns an error if the context
// is canceled or times out.
func hello(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	fmt.Println("server: hello handler started")
	defer fmt.Println("server: hello handler ended")

	// Create a new external request with the same context to ensure that
	// cancellation is propagated to the external request.
	externalReq, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/delay/10", nil)
	if err != nil {
		http.Error(w, "Failed to create external request", http.StatusInternalServerError)
		return
	}

	// Initialize an HTTP client with a timeout to avoid hanging indefinitely.
	// The timeout should generally be less than or equal to the external API's expected response time.
	client := &http.Client{
		Timeout: 15 * time.Second, // Ensures the request doesn't take longer than 15 seconds
	}

	// Create a single-buffered channel to receive the response or error.
	// A buffered channel prevents the goroutine from blocking if the main routine exits early.
	resultChan := make(chan result, 1)

	// Start a goroutine to perform the external HTTP request.
	// Using a goroutine allows the main handler to remain responsive to context cancellation.
	go func() {
		// Perform the external request. The request respects the context's cancellation.
		resp, err := client.Do(externalReq)
		// Send the result to the channel. Since the channel is buffered, this won't block.
		resultChan <- result{resp: resp, err: err}
	}()

	// Use a select statement to wait for either the external request to complete
	// or the context to be canceled/timed out.
	select {
	case res := <-resultChan:
		// External request completed first
		if res.err != nil {
			// Log the error for server-side debugging
			fmt.Println("server:", res.err)
			// Respond to the client with an internal server error
			http.Error(w, res.err.Error(), http.StatusInternalServerError)
			return
		}
		// Ensure the response body is closed to prevent resource leaks
		defer res.resp.Body.Close()
		// For demonstration, write the status code back to the client
		fmt.Fprintf(w, "Received response with status: %s\n", res.resp.Status)

	case <-ctx.Done():
		// Context was canceled or timed out before the external request completed
		err := ctx.Err()
		fmt.Println("server:", err)
		// Respond to the client with the context's error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		// No need to explicitly cancel the external request here because
		// the context cancellation has already propagated to the external request
	}

	// Note:
	// - The buffered channel ensures that the goroutine sending to resultChan
	//   does not block even if the handler returns early due to context cancellation.
	// - Using http.NewRequestWithContext ensures that the external HTTP request
	//   respects the context's cancellation, preventing unnecessary work.
	// - The HTTP client's timeout provides an additional safeguard against hanging requests.
}

func main() {
	// Register the hello handler for the /hello endpoint
	http.HandleFunc("/hello", hello)
	fmt.Println("Server is listening on :8090")
	// Start the HTTP server. In a production environment, you should handle potential errors here.
	http.ListenAndServe(":8090", nil)
}
