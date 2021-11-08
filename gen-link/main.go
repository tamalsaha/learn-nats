package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/tamalsaha/nats-hop-demo/shared"
	"github.com/tamalsaha/nats-hop-demo/transport"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kmodules.xyz/client-go/tools/clusterid"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	auditlib "go.bytebuilders.dev/audit/lib"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/ulids"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2/klogr"
	cu "kmodules.xyz/client-go/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type User struct {
	Name  string
	Email string
}

type LinkData struct {
	LinkID     string
	Token      string
	ClusterID  string
	NotAfter   time.Time
	User       User
	Kubeconfig []byte
}

type UserValues struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type ChartValues struct {
	User UserValues `json:"user"`
}

type VerifyLink struct {
	LinkID    string
	ClusterID string
}

var links = map[string]LinkData{}

const (
	LicenseBucket = "licenses.appscode.com"
	LinkLifetime  = 10 * time.Minute
)

func main() {
	u := User{
		Name:  "Tamal Saha",
		Email: "tamal@appscode.com",
	}
	fs := blobfs.New("gs://" + LicenseBucket)

	/*

		konfig = clientcmd.NewNonInteractiveClientConfig(config, config.CurrentContext, &clientcmd.ConfigOverrides{}, nil)
		return konfig, nil
	*/

	kubeconfigBytes, err := ioutil.ReadFile(filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err)
	}
	if err := genLink(fs, u, kubeconfigBytes); err != nil {
		panic(err)
	}
}

func main_() {
	link := VerifyLink{
		LinkID:    "****",
		ClusterID: "***",
	}
	fs := blobfs.New("gs://" + LicenseBucket)
	nc, err := getNatsClient()
	if err != nil {
		panic(err)
	}
	if err := verifyLink(fs, nc, link); err != nil {
		panic(err)
	}
}

func genLink(fs *blobfs.BlobFS, u User, kubeconfigBytes []byte) error {
	domain := Domain(u.Email)
	now := time.Now()
	timestamp := []byte(now.UTC().Format(time.RFC3339))
	if exists, err := fs.Exists(context.TODO(), EmailVerifiedPath(domain, u.Email)); err == nil && !exists {
		err = fs.WriteFile(context.TODO(), EmailVerifiedPath(domain, u.Email), timestamp)
		if err != nil {
			return err
		}
	}

	token := uuid.New()

	err := fs.WriteFile(context.TODO(), EmailTokenPath(domain, u.Email, token.String()), timestamp)
	if err != nil {
		return err
	}

	linkID := ulids.MustNew().String()

	links[linkID] = LinkData{
		LinkID:     linkID,
		Token:      token.String(),
		ClusterID:  "", // unknown
		NotAfter:   now.Add(LinkLifetime),
		User:       u,
		Kubeconfig: kubeconfigBytes,
	}
	// save link info in the database
	return nil
}

func verifyLink(fs *blobfs.BlobFS, nc *nats.Conn, in VerifyLink) error {
	link, found := links[in.LinkID]
	if !found {
		return fmt.Errorf("unknown link id %q", in.LinkID)
	}
	now := time.Now()
	domain := Domain(link.User.Email)
	if now.After(link.NotAfter) {
		return fmt.Errorf("link %s expired %v ago", link.LinkID, now.Sub(link.NotAfter))
	}

	// check PING
	pong, err := nc.Request(shared.ProxyStatusSubject(in.ClusterID), []byte("PING"), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to ping cluster connector for clustr id %s", in.ClusterID)
	}
	if !bytes.Equal(pong.Data, []byte("PONG")) {
		return fmt.Errorf("expected response PONG from cluster id %s, received %s", in.ClusterID, string(pong.Data))
	}

	// check clusterID
	kc, err := getKubeClient(link.Kubeconfig, nc, in.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to creqate client for clustr id %s, reason: %v", in.ClusterID, err)
	}
	actualClusterID, err := clusterid.ClusterUID(kc.CoreV1().Namespaces())
	if err != nil {
		return fmt.Errorf("failed to read cluster id for cluster %s, reason %v", in.ClusterID, err)
	}
	if in.ClusterID != actualClusterID {
		return fmt.Errorf("actual cluster id %s does not match cluster id %s provided by link %s", actualClusterID, in.ClusterID, in.LinkID)
	}
	link.ClusterID = in.ClusterID

	// store in database cluster_id, kubeconfig for this user

	// delete token
	if exists, err := fs.Exists(context.TODO(), EmailTokenPath(domain, link.User.Email, link.Token)); err == nil && exists {
		err := fs.DeleteFile(context.TODO(), EmailTokenPath(domain, link.User.Email, link.Token))
		if err != nil {
			return err
		}
	}

	// mark link as used ?
	// not needed, since it expires after 10 mins
	// OR
	// keep it private and give users a signed URL?

	return nil
}

func Domain(email string) string {
	idx := strings.LastIndexByte(email, '@')
	if idx == -1 {
		return "_missing_domain_"
	}
	return email[idx+1:]
}

func EmailVerifiedPath(domain, email string) string {
	return fmt.Sprintf("domains/%s/emails/%s/verified", domain, email)
}

func EmailTokenPath(domain, email, token string) string {
	return fmt.Sprintf("domains/%s/emails/%s/tokens/%s", domain, email, token)
}

func getNatsClient() (*nats.Conn, error) {
	var licenseFile string
	flag.StringVar(&licenseFile, "license-file", licenseFile, "Path to license file")
	flag.Parse()

	ctrl.SetLogger(klogr.New())
	config := ctrl.GetConfigOrDie()

	// 	tr, err := cfg.TransportConfig()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	cfg.Transport, err = transport.New(tr, nc, "k8s", 10000*time.Second)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		return nil, err
	}

	c, err := client.New(config, client.Options{
		Scheme: clientgoscheme.Scheme,
		Mapper: mapper,
		Opts: client.WarningHandlerOptions{
			SuppressWarnings:   false,
			AllowDuplicateLogs: false,
		},
	})
	if err != nil {
		return nil, err
	}

	cid, err := cu.ClusterUID(c)
	if err != nil {
		return nil, err
	}

	ncfg, err := auditlib.NewNatsConfig(cid, licenseFile)
	if err != nil {
		return nil, err
	}

	return ncfg.Client, nil
}

func getKubeClient(kubeconfigBytes []byte, nc *nats.Conn, clusterID string) (kubernetes.Interface, error) {
	kubeconfig, err := clientcmd.Load(kubeconfigBytes)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.NewNonInteractiveClientConfig(*kubeconfig, "", &clientcmd.ConfigOverrides{}, nil).ClientConfig()
	if err != nil {
		return nil, err
	}

	//mapper, err := apiutil.NewDynamicRESTMapper(config)
	//if err != nil {
	//	panic(err)
	//}
	//
	//c, err := client.New(config, client.Options{
	//	Scheme: clientgoscheme.Scheme,
	//	Mapper: mapper,
	//	Opts: client.WarningHandlerOptions{
	//		SuppressWarnings:   false,
	//		AllowDuplicateLogs: false,
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}

	c2 := rest.CopyConfig(config)
	cfg, err := c2.TransportConfig()
	if err != nil {
		return nil, err
	}
	c2.Transport, err = transport.New(cfg, nc, shared.ProxyHandlerSubject(clusterID), 10000*time.Second)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c2)
}
