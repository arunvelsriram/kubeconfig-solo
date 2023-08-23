package main

import (
	"fmt"
	"os/exec"
	"os/user"

	"os"

	"gopkg.in/yaml.v3"
)

type Cluster struct {
	Env     string
	Name    string
	Project string
	Context string
	Region  string
}

type Clusters []Cluster

func main() {
	configfile := os.Args[1]
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

	for _, cluster := range clusters {
		fmt.Printf("\n***** %s STARTED *****\n", cluster.Name)
		kubeconfigFile := fmt.Sprintf("%s/%s/%s.yaml", kubeconfigsDir, cluster.Env, cluster.Context)

		if _, err := os.Stat(kubeconfigFile); !os.IsNotExist(err) {
			fmt.Printf("removing existing file: %s\n", kubeconfigFile)
			if err := os.Remove(kubeconfigFile); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

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

		defaultContextName := fmt.Sprintf("gke_%s_%s_%s", cluster.Project, cluster.Region, cluster.Name)
		fmt.Printf("renaming default context name '%s' to '%s'", defaultContextName, cluster.Context)
		cmd = exec.Command("kubectl", "--kubeconfig", kubeconfigFile, "config", "rename-context", defaultContextName, cluster.Context)
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
