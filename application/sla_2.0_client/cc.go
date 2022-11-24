package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func int32Ptr(i int32) *int32 { return &i }

func runCommand(args, env []string, dir *string) (bytes.Buffer, error) {
	if len(args) < 1 {
		return bytes.Buffer{}, fmt.Errorf("too few arguments to run command")
	}

	cmd := exec.Command(args[0], args[1:]...)

	if dir != nil {
		cmd.Dir = *dir
	}

	var outb, errb bytes.Buffer
	outWriter := io.MultiWriter(&outb, os.Stdout)
	errWriter := io.MultiWriter(&errb, os.Stderr)

	cmd.Stdout = outWriter
	cmd.Stderr = errWriter

	if len(env) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, env...)
	}

	err := cmd.Run()
	if err != nil {
		return bytes.Buffer{}, err
	}

	return outb, nil
}

func DeployCC(ccName string, orgNr int, conf lib.Config) error {
	ccLabel := ccName

	// TODO: Change this when changing the registry
	ccImage := "147.102.19.6/pledger/slasc-bridge"

	orgNamespace := "pledger-dlt"

	dir, err := os.MkdirTemp("", "*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	ccPackage := filepath.Join(dir, fmt.Sprintf("%s.tgz", ccName))

	err = packageCC(ccName, ccLabel, ccPackage)
	if err != nil {
		return err
	}

	ccID, err := setId(ccPackage)
	if err != nil {
		return err
	}
	ccID = strings.TrimSpace(ccID)
	log.Printf("Chaincode ID: %s", ccID)

	log.Printf("Deploying Chaincodes")
	err = launchCC(orgNr, orgNamespace, "peer1", ccName, ccID, ccImage)
	if err != nil {
		return err
	}

	err = launchCC(orgNr, orgNamespace, "peer2", ccName, ccID, ccImage)
	if err != nil {
		return err
	}
	log.Printf("Deployed Chaincodes")

	log.Printf("Activating Chaincodes")
	err = ActivateCC(orgNr, ccName, ccPackage, ccID, conf)
	if err != nil {
		return err
	}
	log.Printf("Activated Chaincodes")

	err = queryMetadata(orgNr, ccName, conf)
	if err != nil {
		return err
	}
	return nil
}

func packageCC(ccName, ccLabel, ccPackage string) error {

	args := []string{"./scripts/cc.sh", "package", ccName, ccLabel, ccPackage}

	_, err := runCommand(args, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func setId(ccPackage string) (string, error) {
	args := []string{"./scripts/cc.sh", "id", ccPackage}

	out, err := runCommand(args, nil, nil)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func launchCC(orgNr int, orgNamespace, peer, ccName, ccID, ccImage string) error {
	deploymentName := fmt.Sprintf("org%d%s-ccaas-%s", orgNr, peer, ccName)

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	deploymentsClient := clientset.AppsV1().Deployments(orgNamespace)
	servicesClient := clientset.CoreV1().Services(orgNamespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "main",
							Image:           ccImage,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Env: []apiv1.EnvVar{
								{
									Name:  "CHAINCODE_SERVER_ADDRESS",
									Value: "0.0.0.0:8999",
								},
								{
									Name:  "CHAINCODE_ID",
									Value: ccID,
								},
								{
									Name:  "CORE_CHAINCODE_ID_NAME",
									Value: ccID,
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 8999,
								},
							},
						},
					},
				},
			},
		},
	}

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:     "chaincode",
					Port:     8999,
					Protocol: apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": deploymentName,
			},
		},
	}

	log.Println("Creating deployment...")
	deploymentRes, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("Created deployment %q.\n", deploymentRes.GetObjectMeta().GetName())

	serviceRes, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("Created service %q.\n", serviceRes.GetObjectMeta().GetName())

	return nil
}

func ActivateCC(orgNr int, ccName, ccPackage, ccID string, conf lib.Config) error {
	log.Printf("Installing chaincodes")
	err := installCC(orgNr, 1, ccName, ccPackage, conf)
	if err != nil {
		return err
	}

	err = installCC(orgNr, 2, ccName, ccPackage, conf)
	if err != nil {
		return err
	}

	//QueryInstalled(orgNr, ccName, conf)
	log.Printf("Installed chaincodes")
	log.Printf("Approving chaincodes")

	err = approveCC(orgNr, ccName, ccID, conf)
	if err != nil {
		return err
	}
	log.Printf("Approved chaincodes")

	err = checkCommitReadiness(orgNr, ccName, conf)
	if err != nil {
		return err
	}

	log.Printf("Committing chaincodes")

	err = commitCC(orgNr, ccName, conf)
	if err != nil {
		return err
	}
	log.Printf("Committed chaincodes")

	return nil
}

func installCC(orgNr, peerNr int, ccName, ccPackage string, conf lib.Config) error {
	args := []string{
		"./scripts/cc.sh",
		"install",
		strconv.Itoa(orgNr),
		strconv.Itoa(peerNr),
		conf.TlsCertPath,
		ccPackage,
	}

	_, err := runCommand(args, nil, nil)
	if err != nil {
		return err
	}
	return nil

}

func approveCC(orgNr int, ccName, ccID string, conf lib.Config) error {
	env := []string{
		"CHANNEL_NAME=" + conf.ChannelName,
		"ORDERER_TIMEOUT=10s",
	}

	args := []string{
		"./scripts/cc.sh",
		"approve",
		strconv.Itoa(orgNr),
		ccName,
		ccID,
		conf.TlsCertPath,
	}

	_, err := runCommand(args, env, nil)
	if err != nil {
		return err
	}
	return nil
}

func commitCC(orgNr int, ccName string, conf lib.Config) error {
	env := []string{
		"CHANNEL_NAME=" + conf.ChannelName,
		"ORDERER_TIMEOUT=10s",
	}

	args := []string{
		"./scripts/cc.sh",
		"commit",
		strconv.Itoa(orgNr),
		ccName,
		conf.TlsCertPath,
	}

	_, err := runCommand(args, env, nil)
	if err != nil {
		return err
	}

	return nil
}

func queryMetadata(orgNr int, ccName string, conf lib.Config) error {

	env := []string{
		"CHANNEL_NAME=" + conf.ChannelName,
	}
	args := []string{
		"./scripts/cc.sh",
		"query",
		"metadata",
		ccName,
		strconv.Itoa(orgNr),
		conf.TlsCertPath,
	}

	_, err := runCommand(args, env, nil)
	if err != nil {
		return err
	}

	return nil
}

func QueryInstalled(orgNr int, ccName string, conf lib.Config) (bool, error) {

	args := []string{
		"./scripts/cc.sh",
		"query",
		"installed",
		strconv.Itoa(orgNr),
		conf.TlsCertPath,
	}

	res, err := runCommand(args, nil, nil)
	if err != nil {
		return false, err
	}

	match, _ := regexp.MatchString(ccName, res.String())

	return match, nil
}

func checkCommitReadiness(orgNr int, ccName string, conf lib.Config) error {
	env := []string{
		"CHANNEL_NAME=" + conf.ChannelName,
		"ORDERER_TIMEOUT=10s",
	}

	args := []string{
		"./scripts/cc.sh",
		"checkcommitreadiness",
		strconv.Itoa(orgNr),
		ccName,
		conf.TlsCertPath,
	}

	_, err := runCommand(args, env, nil)
	if err != nil {
		return err
	}

	return nil
}
