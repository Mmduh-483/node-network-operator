package nodenetworkconfigurationpolicy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	k8sv1alpha1 "github.com/pliurh/node-network-operator/pkg/apis/k8s/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nodenetworkconfigurationpolicy")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new NodeNetworkConfigurationPolicy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNodeNetworkConfigurationPolicy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nodenetworkconfigurationpolicy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NodeNetworkConfigurationPolicy
	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.NodeNetworkConfigurationPolicy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner NodeNetworkConfigurationPolicy
	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.NodeNetworkConfigurationPolicy{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &k8sv1alpha1.NodeNetworkConfigurationPolicy{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNodeNetworkConfigurationPolicy{}

// ReconcileNodeNetworkConfigurationPolicy reconciles a NodeNetworkConfigurationPolicy object
type ReconcileNodeNetworkConfigurationPolicy struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NodeNetworkConfigurationPolicy object and makes changes based on the state read
// and what is in the NodeNetworkConfigurationPolicy.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNodeNetworkConfigurationPolicy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NodeNetworkConfigurationPolicy")

	// Fetch the NodeNetworkConfigurationPolicy instance
	instance := &k8sv1alpha1.NodeNetworkConfigurationPolicy{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and  don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	k8sv1alpha1.ValidateNodeNetworkConfigurationPolicy(instance)

	b := new(bytes.Buffer)
    for key, value := range instance.Labels {
        fmt.Fprintf(b, "%s=%s", key, value)
    }
	reqLogger.Info("Instance","Labels", b.String())

	// Fetch the NodeNetworkConfigurationPolicy instances with the same label.
	policies := &k8sv1alpha1.NodeNetworkConfigurationPolicyList{}
	listOpts := &client.ListOptions{}
	listOpts.SetLabelSelector(b.String())
	err = r.client.List(context.TODO(), listOpts, policies)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	policy := k8sv1alpha1.MergeNodeNetworkConfigurationPolicies(policies)
	generateIgnConfig(policy)

	// Config interface type, total number of vfs, and enable sriov
	for _, iface := range instance.Spec.DesiredState.Interfaces {
                err = configPf(iface)
		if err != nil {
			return reconcile.Result{}, err
		}
        }


	err = r.updateNodeNetworkState(policy)
	return reconcile.Result{}, nil
}

func newNodeNetworkState(node *corev1.Node)*k8sv1alpha1.NodeNetworkState{
	return &k8sv1alpha1.NodeNetworkState{
		ObjectMeta: metav1.ObjectMeta{
			Name: node.Name,
		},
		Spec: k8sv1alpha1.NodeNetworkStateSpec{
			Managed: true,
		},
	}
}

func (r *ReconcileNodeNetworkConfigurationPolicy)updateNodeNetworkState(cr *k8sv1alpha1.NodeNetworkConfigurationPolicy) error{
	listOpts := &client.ListOptions{}
	nodes := &corev1.NodeList{}
	err := r.client.List(context.TODO(), listOpts, nodes)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil
		}
		// Error reading the object - requeue the request.
		return err
	}

	for _, node := range nodes.Items {
		cfg := &k8sv1alpha1.NodeNetworkState{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: node.Name, Namespace: ""}, cfg)
		if err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, create it
				log.Info("NodeNetworkState is not found, create it", "node", node.Name)
				cfg := newNodeNetworkState(&node)
				err = r.client.Create(context.TODO(), cfg)
				if err != nil {
					return err
				}
			} else {
				// Error reading the object - requeue the request.
				return err
			}
		}

		// update node network desired config
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: node.Name, Namespace: ""}, cfg)
		if err != nil {
			return err
		}
		log.Info("Update node network config", "desired state", cr.Spec.DesiredState)
		cfg.Status.DesiredState = *cr.Spec.DesiredState.DeepCopy()
		err = r.client.Status().Update(context.TODO(), cfg)
		if err != nil {
			// Error reading the object - requeue the request.
			return err
		}
	}
	return nil
}

func generateIgnConfig(cr *k8sv1alpha1.NodeNetworkConfigurationPolicy) {
	contents := make(map[string]string)
	for _, iface := range cr.Spec.DesiredState.Interfaces {
		parseInterface(iface, contents)
	}
	generateFiles(contents)
}

func parseInterface(i k8sv1alpha1.Interface, contents map[string]string) {
	if i.Mtu != nil && *i.Mtu != 0 {
		contents["mtu"] += fmt.Sprintf("ACTION==\"add\", SUBSYSTEM==\"net\", KERNEL==\"%s\", RUN+=\"/sbin/ip link set mtu %d dev '%%k'\"\n", i.Name, *i.Mtu)
	}
	if i.NumVfs != nil && *i.NumVfs >= 0 {
		link := fmt.Sprintf("/sys/class/net/%s/device", i.Name)
		pciDevDir, err  := os.Readlink(link)
		if err != nil {
			log.Error(err, "failed to get pci bus")
		} else {
			contents["sriov"] += fmt.Sprintf("ACTION==\"add\", SUBSYSTEM==\"pci\", KERNEL==\"%s\", ATTR{sriov_numvfs}=\"%d\"\n", pciDevDir[9:], *i.NumVfs)
		}
		numvfsFilePath := fmt.Sprintf("/sys/class/net/%s/device/sriov_numvfs", i.Name)
		log.Info("file path", "path to numvfs", numvfsFilePath)
		if f,err := os.OpenFile(numvfsFilePath, os.O_WRONLY, 0644); err == nil {
			_,err := f.Write(([]byte)("0"))
			log.Info("err msg", "err:",err)
			f.Close()
			f, _ := os.OpenFile(numvfsFilePath, os.O_WRONLY, 0644)
			f.Write(([]byte)(fmt.Sprintf("%d", *i.NumVfs)))
			f.Close()
	}}
	if i.Promisc != nil {
		filename := "ifcfg-" + i.Name
		contents[filename] += fmt.Sprintf("DEVICE=%s\n", i.Name)
		contents[filename] += fmt.Sprintf("ONBOOT=%s\n", "yes")
		contents[filename] += "NM_CONTROLLED=yes\n"
		if *i.Promisc {
			contents[filename] += fmt.Sprintf("PROMISC=%s\n", "yes")
		} else {
			contents[filename] += fmt.Sprintf("PROMISC=%s\n", "no")
		}
	}
}

func generateFiles(contents map[string]string) {
	mtuFileName,sriovFileName:="/etc/udev/rules.d/99-mtu.rules","/etc/udev/rules.d/99-sriov.rules"
	createFileIfNotExist(mtuFileName)
	createFileIfNotExist(sriovFileName)
	mtuFile, err := os.OpenFile(mtuFileName, os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err, "failed to open 99-mtu.rules")
	}
	defer mtuFile.Close()
	sriovFile, err := os.OpenFile(sriovFileName, os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err, "failed to open 99-sriov.rules")
	}
	defer sriovFile.Close()
	r := regexp.MustCompile(`^ifcfg-.*`)

	for k, v := range contents {
		switch k {
		case "mtu":
			log.Info("file content", "99-mtu.rules", v)
			_, err = mtuFile.Write(([]byte)(v))
			if err != nil{
				log.Error(err, "Failed to write to mtu file")
			}
		case "sriov":
			log.Info("file content", "99-sriov.rules", v)
			_, err = sriovFile.Write(([]byte)(v))
			if err != nil{
                                log.Error(err, "Failed to write to sriov file")
                        }
		default:
			if r.MatchString(k) {
				log.Info("file content", k, v)
				path:= fmt.Sprintf("/etc/sysconfig/network-scripts/%s", k)
				createFileIfNotExist(path)
				f, err := os.OpenFile(path, os.O_WRONLY, 0644)
				if err != nil{
					log.Error(err, "failed to open ifcfg-file rules")
				}
				defer f.Close()
				f.Write(([]byte)(v))
			}
		}
	}
}

func createFileIfNotExist(path string)  {
	_, err := os.Stat(path)
	if err != nil {
		os.Create(path)
	}
}

func configPf(iface k8sv1alpha1.Interface) error {
	log.Info("Configpf", "PF",iface.Name)
	log.Info("Configpf", "iface", iface)
        args := make(map[string]string, 0)
        args["SRIOV_EN"] = "True"
        args["LINK_TYPE_P1"] = "2"
        args["NUM_OF_VFS"] = fmt.Sprintf("%d", iface.TotalVfs)
        pciAddrs, err := getPciAddrsFromNetDevName(iface.Name)
        if err != nil {
                return nil
        }
	log.Info("Configpf", "PCIADDRS",pciAddrs)
        for k, v := range args {
		log.Info("Configpf", k , v)
                command := fmt.Sprintf("mstconfig -y -d %s set %s=%s", pciAddrs, k, v)
                cmd := exec.Command("bash", "-c",command)
                var out bytes.Buffer
                cmd.Stdout = &out
                err = cmd.Run()
                if err != nil {
                        return err
                }
        }
	return nil
}

func getPciAddrsFromNetDevName(ifname string) (string, error){
        link := "/sys/class/net/" + ifname + "/device"
        pciDevDir, err := os.Readlink(link)
        if err != nil ||  len(pciDevDir) <= 9 {
                return "", fmt.Errorf("could not find PCI Address for net dev %s", ifname)
        }
        return pciDevDir[9:], nil
}

