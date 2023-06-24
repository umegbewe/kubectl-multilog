package multilog

import (
	"bufio"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spaolacci/murmur3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var colorPool = []color.Attribute{
	color.FgCyan,
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
}

var colorMap = map[string]func(...interface{}) string{}

func StreamLogs(ctx context.Context, logger *logrus.Logger, kubeconfig string, kubeContext string, namespace string, selector string, initContainers bool, previous bool, tailLines int64) error {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	overrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %v", err)
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("error listing pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found matching selector %s", selector)
	}

	logger.Infof("Found %d pod(s)", len(pods.Items))

	var wg sync.WaitGroup
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			wg.Add(1)
			streamLogger := logger.WithFields(logrus.Fields{
				"pod":       pod.Name,
				"namespace": pod.Namespace,
				"container": container.Name,
			})
			go streamContainerLogs(ctx, streamLogger, clientset, pod, container.Name, previous, tailLines, &wg)
		}
		if initContainers {
			for _, container := range pod.Spec.InitContainers {
				streamLogger := logger.WithFields(logrus.Fields{
					"pod":       pod.Name,
					"namespace": pod.Namespace,
					"container": container.Name,
				})
				go streamContainerLogs(ctx, streamLogger, clientset, pod, container.Name, previous, tailLines, &wg)
			}
		}
	}

	wg.Wait()
	return nil
}

func streamContainerLogs(ctx context.Context, logger *logrus.Entry, clientset *kubernetes.Clientset, pod corev1.Pod, container string, previous bool, tailLines int64, wg *sync.WaitGroup) {
	defer wg.Done()
	logger = logger.WithFields(logrus.Fields{
		"pod":       pod.Name,
		"namespace": pod.Namespace,
		"container": container,
	})

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container,
		Follow:    true,
		Previous:  previous,
		TailLines: &tailLines,
	})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Errorf("pod not found: %v", err)
			return
		}
	}

	defer podLogs.Close()

	colorFunc := getColorFuncForPod(pod.Name, container)

	scanner := bufio.NewScanner(podLogs)

	for scanner.Scan() {
		prefix := fmt.Sprintf("[pod=%s][namespace=%s][container=%s] %s", pod.Name, pod.Namespace, container, scanner.Text())
		fmt.Println(colorFunc(prefix))
	}

	//copyAndClose := func(dst io.Writer, src io.ReadCloser) {
	//	copyDone := make(chan struct{})
	//	go func() {
	//		io.Copy(dst, src)
	//		close(copyDone)
	//	}()
	//	<-copyDone
	//	src.Close()
	//}
	//copyAndClose(os.Stdout, podLogs)
}

func getColorFuncForPod(pod string, containerName string) func(...interface{}) string {
	key := fmt.Sprintf("%s-%s", pod, containerName) // unique key for pod/container combo
	if colorFunc, ok := colorMap[key]; ok {
		return colorFunc
	}

	colorIndex := int(murmur3.Sum32WithSeed([]byte(key), 0) % uint32(len(colorPool)))

	colorFunc := color.New(colorPool[colorIndex]).SprintFunc()
	return colorFunc
}
