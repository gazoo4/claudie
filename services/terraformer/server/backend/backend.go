package backend

import (
	"fmt"

	"github.com/Berops/claudie/internal/envs"
	"github.com/Berops/claudie/internal/templateUtils"
)

var (
	minioURL  = envs.MinioURL
	accessKey = envs.MinioAccessKey
	secretKey = envs.MinioSecretKey
)

type Backend struct {
	ProjectName string
	ClusterName string
	Directory   string
}

type templateData struct {
	ProjectName string
	ClusterName string
	MinioURL    string
	AccessKey   string
	SecretKey   string
}

// function CreateFiles will create a backend.tf file from template
func (b Backend) CreateFiles() error {
	template := templateUtils.Templates{Directory: b.Directory}
	templateLoader := templateUtils.TemplateLoader{Directory: templateUtils.TerraformerTemplates}
	tpl, err := templateLoader.LoadTemplate("backend.tpl")
	if err != nil {
		return fmt.Errorf("error while parsing template file backend.tpl: %v", err)
	}
	data := templateData{ProjectName: b.ProjectName, ClusterName: b.ClusterName, MinioURL: minioURL, AccessKey: accessKey, SecretKey: secretKey}
	err = template.Generate(tpl, "backend.tf", data)
	if err != nil {
		return fmt.Errorf("error while creating backend files: %v", err)
	}
	return nil
}
