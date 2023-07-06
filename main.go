package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"topwatcher/pkg/reader"
)

type Info struct {
	Deployment string
	Kind       string
	Replicas   int
	Pods       []map[string]string
}

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
	exceptions    []string
	configFile    reader.Configuration
)

func init() {
	var flags int

	//	path := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	configFile = reader.ReadFile()
	if configFile.Logging.Debug {
		flags = log.Ldate | log.Ltime | log.Lshortfile
		DebugLogger = log.New(os.Stdout, "DEBUG ", flags)
	} else {
		flags = log.Ldate | log.Ltime
	}
	InfoLogger = log.New(os.Stdout, "INFO ", flags)
	WarningLogger = log.New(os.Stdout, "WARNING ", flags)
	ErrorLogger = log.New(os.Stdout, "ERROR ", flags)

	InfoLogger.Println("Starting topwatcher...")

	if configFile.Logging.Debug {
		DebugLogger.Println("Reading Configuration file...")
	}

	allkeys := make(map[string]bool)

	for _, item := range configFile.Kubernetes.Exceptions.Deployments {
		if _, value := allkeys[item]; !value {
			allkeys[item] = true
			exceptions = append(exceptions, item)
		}
	}
}

func main() {

	fmt.Println(configFile.Kubernetes.Namespaces)
	var alerts []string
	var target []string

	clientSet, config := GetClusterAccess(&configFile)

	if len(configFile.Kubernetes.Namespaces) > 0 {
		if Contain(configFile.Kubernetes.Namespaces, clientSet) {
			podInfo := GetPodInfo(clientSet, &configFile, config)
			if configFile.Logging.Debug {
				DebugLogger.Printf("Pods information list is: %v", podInfo)
			}
			if configFile.Kubernetes.Threshold.Ram > 0 {
				alerts, target = CheckPodRamUsage(&configFile, podInfo)
			} else {
				ErrorLogger.Println("Ram value is not defined in configuration file")
				os.Exit(1)
			}
		} else {
			WarningLogger.Printf("'%v' namespace is not in the cluster!!", configFile.Kubernetes.Namespaces)
		}
	} else {
		ErrorLogger.Println("Namespace is not defined")
		os.Exit(1)
	}

	if len(target) > 0 && configFile.Kubernetes.PodRestart {
		RestartDeployment(clientSet, target)
	}

	if configFile.Slack.Notify && len(configFile.Slack.Channel) > 0 {
		SendSlackPayload(&configFile, alerts)
	} else {
		for _, alert := range alerts {
			InfoLogger.Println(alert)
		}
	}
}
