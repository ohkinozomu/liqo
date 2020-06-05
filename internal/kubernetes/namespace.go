package kubernetes

import (
	"errors"
	nattingv1 "github.com/liqoTech/liqo/api/namespaceNattingTable/v1"
	"github.com/liqoTech/liqo/pkg/crdClient"
	"github.com/liqoTech/liqo/pkg/crdClient/v1alpha1"
	v1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"strings"
)

type namespaceNTCache struct {
	Store            cache.Store
	Controller       chan struct{}
	nattingTableName string
}

func (p *KubernetesProvider) startNattingCache(clientSet v1alpha1.NamespacedCRDClientInterface) {
	ehf := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			p.StartReflector()
			p.manageReflections(nil, obj)
		},
		UpdateFunc: p.manageReflections,
		DeleteFunc: func(obj interface{}) {
			p.StopReflector()
			if err := p.createNattingTable(p.foreignClusterId); err != nil {
				p.log.Error(err, "cannot create nattingTable")
			}
		},
	}
	lo := metav1.ListOptions{FieldSelector: strings.Join([]string{"metadata.name", p.ntCache.nattingTableName}, "=")}

	p.ntCache.Store, p.ntCache.Controller = crdClient.WatchResources(clientSet,
		"namespacenattingtables", "",
		0, ehf, lo)
}

func (nt *namespaceNTCache) WaitNamespaceNattingTableSync() {
	cache.WaitForCacheSync(nt.Controller, func() bool {
		_, exists, _ := nt.Store.GetByKey(nt.nattingTableName)

		if exists {
			return true
		}

		return false
	})
}

func (nt *namespaceNTCache) getNattingTable(nattingTableName string) (*nattingv1.NamespaceNattingTable, error) {
	o, exists, err := nt.Store.GetByKey(nattingTableName)

	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return o.(*nattingv1.NamespaceNattingTable), nil
}

func (p *KubernetesProvider) NatNamespace(namespace string, create bool) (string, error) {
	nt, exists, err := p.ntCache.Store.GetByKey(p.foreignClusterId)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", errors.New("namespacenattingtable not existing")
	}

	nattingTable := nt.(*nattingv1.NamespaceNattingTable)
	nattedNS, ok := nattingTable.Spec.NattingTable[namespace]
	if !ok && !create {
		return "", errors.New("not natted namespaces")
	}

	if !ok && create {
		nattedNS = strings.Join([]string{p.homeClusterID, namespace}, "-")
		if nattingTable.Spec.NattingTable == nil {
			nattingTable.Spec.NattingTable = make(map[string]string)
			nattingTable.Spec.DeNattingTable = make(map[string]string)
		}

		nattingTable.Spec.NattingTable[namespace] = nattedNS
		nattingTable.Spec.DeNattingTable[nattedNS] = namespace

		_, err := p.homeClient.Resource("namespacenattingtables").Update(nattingTable.Name, nattingTable, metav1.UpdateOptions{})
		if err != nil {
			return "", err
		}

		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nattedNS,
			},
		}

		_, err = p.foreignClient.Client().CoreV1().Namespaces().Create(ns)
		if err != nil && !kerror.IsAlreadyExists(err) {
			return "", err
		}
	}

	return nattedNS, nil
}

func (p *KubernetesProvider) DeNatNamespace(namespace string) (string, error) {
	nt, exists, err := p.ntCache.Store.GetByKey(p.foreignClusterId)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", errors.New("namespacenattingtable not existing")
	}

	deNattedNS, ok := nt.(*nattingv1.NamespaceNattingTable).Spec.DeNattingTable[namespace]
	if !ok {
		return "", errors.New("not natted namespaces")
	}

	return deNattedNS, nil
}

func (p *KubernetesProvider) createNattingTable(name string) error {
	_, err := p.homeClient.Resource("namespacenattingtables").Get(name, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if !kerror.IsNotFound(err) {
		return err
	}

	var table runtime.Object

	table = &nattingv1.NamespaceNattingTable{
		TypeMeta: metav1.TypeMeta{
			Kind: "NamespaceNattingTable",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: nattingv1.NamespaceNattingTableSpec{
			ClusterId:    name,
			NattingTable: map[string]string{},
		},
	}

	_, err = p.homeClient.Resource("namespacenattingtables").Create(table, metav1.CreateOptions{})

	if err != nil && kerror.IsAlreadyExists(err) {
		return nil
	}

	return err
}

func (p *KubernetesProvider) manageReflections(oldObj interface{}, newObj interface{}) {
	var oldNt map[string]string

	nt := newObj.(*nattingv1.NamespaceNattingTable).Spec.NattingTable
	if oldObj != nil {
		oldNt = oldObj.(*nattingv1.NamespaceNattingTable).Spec.NattingTable
	}

	p.reflectedNamespaces.Lock()
	defer p.reflectedNamespaces.Unlock()

	for k, v := range nt {
		if _, ok := p.reflectedNamespaces.ns[k]; !ok {

			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: v,
				},
			}

			_, err := p.foreignClient.Client().CoreV1().Namespaces().Create(ns)
			if err != nil && !kerror.IsAlreadyExists(err) {
				p.log.Error(err, "error in namespace creation")
				continue
			}

			if err := p.reflectNamespace(k); err != nil {
				p.log.Error(err, "error in manage reflections")
				continue
			}
		}
	}

	for k, v := range p.reflectedNamespaces.ns {
		if _, ok := nt[k]; !ok {

			close(v)
			if r := recover(); r != nil {
				p.log.Info("channel already closed by the reflection routine")
			} else {
				if err := p.cleanupNamespace(oldNt[k]); err != nil {
					p.log.Error(err, "error in cleaning up namespace")
				}
			}
			delete(p.reflectedNamespaces.ns, k)
		}
	}
}