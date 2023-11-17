package collect

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ADD    = "ADD"
	UPDATE = "UPDATE"
	DELETE = "DELETE"

	noResync = 0
)

func Secrets() (map[string]*v1.Secret, error) {
	errorCh := make(chan error, 1)
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)

	secrets := map[string]*v1.Secret{}
	go handleEvents(errorCh, secrets)

	select {
	case <-interruptCh:
		return secrets, nil
	case err := <-errorCh:
		return nil, err
	}
}

func handleEvents(errorCh chan error, secrets map[string]*v1.Secret) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		errorCh <- err
		return
	}

	sif := informers.NewSharedInformerFactory(clientset, noResync)

	secretInformer := sif.Core().V1().Secrets().Informer()

	stopCh := make(chan struct{})
	go sif.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, secretInformer.HasSynced) {
		panic("timed out wait for caches to sync")
	}

	_, err = secretInformer.AddEventHandler(&eventHandlerImpl{
		secrets: secrets,
	})
	if err != nil {
		errorCh <- err
		return
	}
}

type eventHandlerImpl struct {
	secrets map[string]*v1.Secret
	errorCh chan error
}

func (e *eventHandlerImpl) OnAdd(obj interface{}, isInInitialList bool) {
	switch v := obj.(type) {
	case *v1.Secret:
		// fmt.Printf("[DUMP] %s (inInitialList %t)\n", v.Name, isInInitialList)
		e.OnSecret(ADD, v)
	default:
		Dunno(obj)
	}
}
func (e *eventHandlerImpl) OnUpdate(oldObj, newObj interface{}) {
	switch v := oldObj.(type) {
	case *v1.Secret:
		e.OnSecret(UPDATE, v)
	default:
		Dunno(newObj)
	}
}
func (e *eventHandlerImpl) OnDelete(obj interface{}) {
	switch v := obj.(type) {
	case *v1.Secret:
		e.OnSecret(DELETE, v)
	default:
		Dunno(obj)
	}
}

func (e *eventHandlerImpl) OnSecret(action string, v *v1.Secret) {

	id := string(v.GetUID())
	switch action {
	case ADD:
		_, ok := e.secrets[id]
		if ok {
			e.errorCh <- fmt.Errorf("secret added when already exists, numbers may be off")
			return
		}
		e.secrets[id] = v
	case UPDATE:
		e.secrets[id] = v
	case DELETE:
		delete(e.secrets, id)
	default:
		return
	}

	fmt.Printf("\r%d secrets collected ", len(e.secrets))
}

func Dunno(obj any) {
	fmt.Printf("Dunno: %T %+v\n", obj, obj)
}
