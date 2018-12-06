package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"log"
	"net/http"
	"time"

	"k8s.io/client-go/kubernetes"

	k8sNode "managedkube.com/kubernetes-cost-agent/pkg/metrics/k8s/node"
	k8sPersistentVolume "managedkube.com/kubernetes-cost-agent/pkg/metrics/k8s/persistentVolume"
	k8sPod "managedkube.com/kubernetes-cost-agent/pkg/metrics/k8s/pod"
)

var exportURL = ""
var exportToken = ""

func SetExportURL(url string){
	exportURL = url
}

func SetExportToken(token string){
	exportToken = token
}

func Run(clientset *kubernetes.Clientset) {

	k8sNode.Register()
	k8sPod.Register()
	k8sPersistentVolume.Register()
	//k8sNamespace.Register()

	go k8sNode.Watch(clientset)
	time.Sleep(5 * time.Second)
	go k8sPod.Watch(clientset)
	go k8sPersistentVolume.Watch(clientset)

	if exportURL != "" {
		go export()
	}
}

func export(){
	update()
}

func update(){
	for{
		time.Sleep(10 * time.Second)
		glog.V(3).Infof("Sending exports")

		podList := k8sPod.GetList()

		for _, p := range podList.Pod {
			fmt.Println(p.Container_name)

			bytesRepresentation, err := json.Marshal(p)
			if err != nil {
				log.Fatalln(err)
			}

			go send(bytesRepresentation)
		}

	}
}

func send(bytesRepresentation []uint8) {

	//resp, err := http.Post(exportURL, "application/json", bytes.NewBuffer(bytesRepresentation))
	//if err != nil {
	//	log.Fatalln(err)
	//}

	timeout := time.Duration(5 * time.Second)

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", exportURL, bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Apikey", exportToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	//var result map[string]interface{}
	//
	//json.NewDecoder(resp.Body).Decode(&result)
	//
	//log.Println(result)
	//log.Println(result["data"])

	if resp.StatusCode != 200 {
		glog.V(3).Infof("Error sending export to: %s, StatusCode: %s", exportURL, resp.Status)
	}
}