package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"gopkg.in/yaml.v3"
)

type Cluster struct {
	Env     string
	Name    string
	Project string
	Context string
	Region  string
	Type    string
}

type Clusters []Cluster

func main() {
	env := flag.String("e", "", "create kubeconfigs for clusters belonging to given env")
	clusterName := flag.String("c", "", "create kubeconfigs for the given cluster name only")

	flag.Parse()

	configfile := flag.Args()[0]
	fmt.Printf("config file: %s\n", configfile)

	f, err := os.ReadFile(configfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var clusters Clusters
	yaml.Unmarshal(f, &clusters)

	user, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	homeDir := user.HomeDir
	kubeconfigsDir := fmt.Sprintf("%s/.kube/configs", homeDir)

	var defaultContextName string

	for _, cluster := range clusters {
		if len(*env) != 0 && cluster.Env != *env {
			fmt.Printf("skipping %s as it does not belong to given env %s\n", cluster.Name, *env)
			continue
		}

		if len(*clusterName) != 0 && cluster.Name != *clusterName {
			fmt.Printf("skipping %s as it does not match the given cluster name %s\n", cluster.Name, *clusterName)
			continue
		}

		fmt.Printf("\n***** %s STARTED *****\n", cluster.Name)

		kubeconfigFile := fmt.Sprintf("%s/%s/%s.yaml", kubeconfigsDir, cluster.Env, cluster.Context)

		if _, err := os.Stat(kubeconfigFile); !os.IsNotExist(err) {
			fmt.Printf("removing existing file: %s\n", kubeconfigFile)
			if err := os.Remove(kubeconfigFile); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if strings.EqualFold(cluster.Type, "gke") {
			cmd := exec.Command("gcloud", "container", "clusters", "get-credentials", cluster.Name, "--project", cluster.Project, "--region", cluster.Region)
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Printf("command: %v\n", cmd)
			if err := cmd.Run(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			defaultContextName = fmt.Sprintf("gke_%s_%s_%s", cluster.Project, cluster.Region, cluster.Name)
		}

		if strings.EqualFold(cluster.Type, "kind") {
			cmd := exec.Command("kind", "export", "kubeconfig", "--name", cluster.Name, "--kubeconfig", kubeconfigFile)
			cmd.Env = os.Environ()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Printf("command: %v\n", cmd)
			if err := cmd.Run(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			defaultContextName = fmt.Sprintf("kind-%s", cluster.Name)
		}

		fmt.Printf("renaming default context name '%s' to '%s'", defaultContextName, cluster.Context)
		cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigFile, "config", "rename-context", defaultContextName, cluster.Context)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("command: %v\n", cmd)
		if err := cmd.Run(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("***** %s COMPLETED *****\n\n", cluster.Name)
	}
}
