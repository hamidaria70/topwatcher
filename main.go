package main

import (
	"fmt"
	"log"
	"os"
)

type Configuration struct {
	Kubernetes struct {
		Namespaces string `yaml:"namespaces"`
		Threshold  struct {
			Ram int `yaml:"ram"`
		} `yaml:"threshold"`
		Exceptions struct {
			Deployments []string `yaml:"deployments,flow"`
		} `yaml:"exceptions"`
	} `yaml:"kubernetes"`
	Slack struct {
		WebhookUrl string `yaml:"webhookurl"`
		Notify     bool   `yaml:"notify"`
		Channel    string `yaml:"channel"`
		UserName   string `yaml:"username"`
	} `yaml:"slack"`
	Logging struct {
		Debug bool `yaml:"debug"`
	} `yaml:"logging"`
}

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
)
var configFile Configuration
var exceptions []string

func init() {
	var flags int

	readFile(&configFile)
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
	var alerts []string
	var target []string
	var info Info
	info.Pods = make([]map[string]string, 0)
	var result []Info

	clientSet, config := GetClusterAccess()

	if len(configFile.Kubernetes.Namespaces) > 0 {
		if Contain(configFile.Kubernetes.Namespaces, clientSet) {
			podDetailList, podMetricsDetailList := GetPodInfo(clientSet, configFile, config)
			podInfo := MergePodMetricMaps(podDetailList, podMetricsDetailList)
			keys := make(map[string]int)
			for _, entry := range podDetailList {
				keys[entry["deployment"]]++
			}
			for j, n := range podDetailList {
				if n["name"] == podMetricsDetailList[j]["name"] {
					if info.Deployment != n["deployment"] && info.Deployment != "" {
						info.Pods = nil
					}
					info.Deployment = n["deployment"]
					info.Kind = n["kind"]
					info.Pods = append(info.Pods, podMetricsDetailList[j])

				}
				result = append(result, info)
			}

			for _, v := range result {
				fmt.Println(v)
			}
			if configFile.Logging.Debug {
				DebugLogger.Printf("Pods information list is: %v", podInfo)
			}
			if configFile.Kubernetes.Threshold.Ram > 0 {
				alerts, target = CheckPodRamUsage(configFile, podInfo)
			} else {
				ErrorLogger.Println("Ram value is not defined in configuration file")
			}
		} else {
			WarningLogger.Printf("'%v' namespace is not in the cluster!!", configFile.Kubernetes.Namespaces)
		}
	} else {
		ErrorLogger.Println("Namespace is not defined")
	}

	if len(target) > 0 {
		RestartDeployment(clientSet, target)
	}

	if configFile.Slack.Notify && len(configFile.Slack.Channel) > 0 {
		SendSlackPayload(configFile, alerts)
	} else {
		for _, alert := range alerts {
			InfoLogger.Println(alert)
		}
	}
}
