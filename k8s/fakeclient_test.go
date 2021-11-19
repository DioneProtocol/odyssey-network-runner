package k8s

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	k8sapi "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/gorilla/mux"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
)

// fakeOperatorClient fakes (mocks) the k8s operator object
// and implements the k8s client.Client interface
type fakeOperatorClient struct {
	nodes []*k8sapi.Avalanchego
	srv   *http.Server
	quit  chan struct{}
	wg    sync.WaitGroup
}

// root serves / HTTP requests
func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

// healthCheck serves the health API endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"healthy":true},"id":1}`))
}

// info serves the info API endpoint
func info(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","result":{"nodeID":"NodeID-%s"},"id":1}`, ids.GenerateTestShortID().String())))
}

// runHTTPServer runs a HTTP server which fakes real avalanchego API calls
func (f *fakeOperatorClient) runHTTPServer() {
	router := mux.NewRouter()
	router.HandleFunc("/ext/health", healthCheck).Methods("POST")
	router.HandleFunc("/ext/info", info).Methods("POST")
	router.HandleFunc("/", root).Methods("GET")

	f.srv = &http.Server{
		Addr:    ":9650",
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := f.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("error on server listen: %s\n", err)
		}
	}()

	f.wg.Add(1)
	select {
	case <-done:
	case <-f.quit:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := f.srv.Shutdown(ctx); err != nil {
		fmt.Printf("error on server shutdown: %s\n", err)
	}
	fmt.Println("HTTP server exited properly")
	f.wg.Done()
}

// newFakeOperatorClient creates a new mock k8s interface
func newFakeOperatorClient() (*fakeOperatorClient, error) {
	f := &fakeOperatorClient{
		nodes: make([]*k8sapi.Avalanchego, 0),
		quit:  make(chan struct{}),
	}
	go f.runHTTPServer()
	return f, nil
}

// Close (so that the HTTP server can shut down)
func (f *fakeOperatorClient) Close() {
	close(f.quit)
	f.wg.Wait()
}

// Scheme implements client.Client
func (f *fakeOperatorClient) Scheme() *runtime.Scheme {
	return nil
}

// RESTMapper implements client.Client
func (f *fakeOperatorClient) RESTMapper() meta.RESTMapper {
	return nil
}

// Get implements client.Client
func (f *fakeOperatorClient) Get(ctx context.Context, key k8scli.ObjectKey, obj k8scli.Object) error {
	for _, n := range f.nodes {
		if n.Name == key.Name && n.Namespace == key.Namespace {
			// obj = n.DeepCopy()
			return nil
		}
	}
	return fmt.Errorf("Couldn't find node %s", key.Name)
}

// List implements client.Client
func (f *fakeOperatorClient) List(ctx context.Context, list k8scli.ObjectList, opts ...k8scli.ListOption) error {
	return nil
}

// Create implements client.Client
func (f *fakeOperatorClient) Create(ctx context.Context, obj k8scli.Object, opts ...k8scli.CreateOption) error {
	var avago *k8sapi.Avalanchego
	var ok bool

	if avago, ok = obj.(*k8sapi.Avalanchego); !ok {
		return fmt.Errorf("Expected Avalanchego object, got %T", obj)
	}
	avago.Status.NetworkMembersURI = []string{"localhost"}
	f.nodes = append(f.nodes, avago)
	return nil
}

// Delete implements client.Client, deletes the given obj from Kubernetes cluster.
func (f *fakeOperatorClient) Delete(ctx context.Context, obj k8scli.Object, opts ...k8scli.DeleteOption) error {
	var avago *k8sapi.Avalanchego
	var ok bool

	if avago, ok = obj.(*k8sapi.Avalanchego); !ok {
		return fmt.Errorf("Expected Avalanchego object, got %T", obj)
	}
	for i, n := range f.nodes {
		if n.Name == avago.Name {
			f.nodes[i] = f.nodes[len(f.nodes)-1]
			f.nodes[len(f.nodes)-1] = nil
			f.nodes = f.nodes[:len(f.nodes)-1]
			return nil
		}
	}
	return fmt.Errorf("Couldn't find node %s", avago.Name)
}

// Update implements client.Client
func (f *fakeOperatorClient) Update(ctx context.Context, obj k8scli.Object, opts ...k8scli.UpdateOption) error {
	return nil
}

// Patch implements client.Client
func (f *fakeOperatorClient) Patch(ctx context.Context, obj k8scli.Object, patch k8scli.Patch, opts ...k8scli.PatchOption) error {
	return nil
}

// DeleteAllOf implements client.Client
func (f *fakeOperatorClient) DeleteAllOf(ctx context.Context, obj k8scli.Object, opts ...k8scli.DeleteAllOfOption) error {
	return nil
}

// Status implements client.Client
func (f *fakeOperatorClient) Status() k8scli.StatusWriter {
	return nil
}
