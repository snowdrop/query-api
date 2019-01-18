package main

import (
	"fmt"
	"github.com/snowdrop/query-api/pkg/helper/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path/filepath"
)

const svcTemplate = `apiVersion: v1
kind: Service
metadata:
  generation: 1
  labels:
    app: my-spring-boot
    name: my-spring-boot
  name: my-spring-boot
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: my-spring-boot
  sessionAffinity: None
  type: ClusterIP
`

type component struct {
	name string
	version string
}

func main() {
	err := NewComponent().CreateChart()
	if (err != nil) {
		panic (err)
	}
}

func NewComponent() *component {
	return &component{}
}

func (c *component) CreateChart() error {
	c.name = "dummy"
	c.version = "v1.0"

	fmt.Printf( "Creating %s chart\n", c.name)
	chartname := filepath.Base(c.name)
	cfile := &chart.Metadata{
		Name:        chartname,
		Description: "A Helm chart for Kubernetes",
		Version:     c.version,
		AppVersion:  "1.0",
		ApiVersion:  helm.ApiVersionV1,
	}

	cpath := filepath.Dir(c.name)
	_, err := helm.Create(cfile, cpath)
	if err != nil {
		return err
	}

	tfile := &chart.Template{
		Name: "templates/service.yml",
		Data: []byte(svcTemplate),
	}

	_, err = helm.SaveTemplate(tfile, c.name)
	if err != nil {
		return err
	}
	return nil
}

