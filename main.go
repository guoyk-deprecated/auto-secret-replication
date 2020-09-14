package main

import (
	"context"
	"fmt"
	"github.com/guoyk93/conc"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	EnvKeySourceNamespace   = "SOURCE_NAMESPACE"
	AnnotationKeyEnabled    = "net.guoyk.auto-secret-replication/enabled"
	AnnotationKeyReplicated = "net.guoyk.auto-secret-replication/replicated"
)

var (
	optDryRun, _ = strconv.ParseBool(os.Getenv("SECRET_AUTO_REPLICATION_DRY_RUN"))

	gConfig          *rest.Config
	gClient          *kubernetes.Clientset
	gSourceNamespace = os.Getenv(EnvKeySourceNamespace)

	gLocker = &sync.Mutex{}
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.Lmsgprefix)
	if optDryRun {
		log.SetPrefix("(dry) ")
	}

	var err error
	defer exit(&err)

	gSourceNamespace = strings.TrimSpace(gSourceNamespace)
	if gSourceNamespace == "" {
		err = fmt.Errorf("missing environment variable: %s", EnvKeySourceNamespace)
		return
	}

	if gConfig, err = rest.InClusterConfig(); err != nil {
		return
	}
	if gClient, err = kubernetes.NewForConfig(gConfig); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		errChan <- conc.Parallel(
			routinePeriodical(),
			routineWatchNamespace(),
			routineWatchSecret(),
		).Do(ctx)
	}()

	select {
	case err = <-errChan:
		return
	case sig := <-sigChan:
		log.Printf("signal caught: %s", sig.String())
		cancel()
		<-errChan
	}

}

func replicateSecrets(ctx context.Context, namespaces []string) (err error) {
	gLocker.Lock()
	defer gLocker.Unlock()

	var sList *corev1.SecretList
	if sList, err = gClient.CoreV1().Secrets(gSourceNamespace).List(ctx, metav1.ListOptions{}); err != nil {
		return
	}
	for _, s := range sList.Items {
		if s.Annotations == nil {
			continue
		}
		if enabled, _ := strconv.ParseBool(s.Annotations[AnnotationKeyEnabled]); !enabled {
			continue
		}
		log.Printf("secret: %s/%s ready to replicate", gSourceNamespace, s.Name)
		for _, namespace := range namespaces {
			// skip source namespace
			if namespace == gSourceNamespace {
				continue
			}
			// create replicated secret
			rs := s.DeepCopy()
			delete(rs.Annotations, AnnotationKeyEnabled)
			rs.Annotations[AnnotationKeyReplicated] = "true"
			rs.Namespace = namespace
			// create or update
			var existed *corev1.Secret
			if existed, err = gClient.CoreV1().Secrets(namespace).Get(ctx, s.Name, metav1.GetOptions{}); err != nil {
				if !k8serrors.IsNotFound(err) {
					return
				}
				log.Printf("secret: %s/%s not found, will create", namespace, s.Name)
				if _, err = gClient.CoreV1().Secrets(namespace).Create(ctx, rs, metav1.CreateOptions{}); err != nil {
					log.Printf("failed to create %s/%s: %s", namespace, rs.Name, err.Error())
					err = nil
					continue
				}
				log.Printf("secret: %s/%s created", namespace, s.Name)
			} else {
				if existed.Annotations == nil {
					log.Printf("missing annotations on existed replicated secrets: %s/%s", namespace, rs.Name)
					continue
				}
				if replicated, _ := strconv.ParseBool(existed.Annotations[AnnotationKeyReplicated]); !replicated {
					log.Printf("missing annotation on existed replicated secrets: %s/%s", namespace, rs.Name)
					continue
				}
				log.Printf("secret: %s/%s already exists, will update", namespace, s.Name)
				if _, err = gClient.CoreV1().Secrets(namespace).Update(ctx, rs, metav1.UpdateOptions{}); err != nil {
					log.Printf("failed to update %s/%s: %s", namespace, rs.Name, err.Error())
					err = nil
					continue
				}
				log.Printf("secret: %s/%s updated", namespace, s.Name)
			}
		}
	}
	return
}

func routinePeriodical() conc.Task {
	return conc.TaskFunc(func(ctx context.Context) (err error) {
		tk := time.NewTicker(time.Minute * 15)
		for {
			log.Println("routine periodical fired")

			// list all namespaces
			var nss *corev1.NamespaceList
			if nss, err = gClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err != nil {
				return
			}

			var namespaces []string
			for _, ns := range nss.Items {
				namespaces = append(namespaces, ns.Name)
			}

			log.Println("found namespaces:", strings.Join(namespaces, ","))

			if err = replicateSecrets(ctx, namespaces); err != nil {
				return
			}

			// wait done or ticker
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
			}
		}
	})
}

func routineWatchNamespace() conc.Task {
	return conc.TaskFunc(func(ctx context.Context) (err error) {
		return
	})
}

func routineWatchSecret() conc.Task {
	return conc.TaskFunc(func(ctx context.Context) (err error) {
		return
	})
}
