apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
name: cluster

versions:
  kubernetes: '{{ .Kubernetes }}'

clusterNetwork:
  cni:
    external: {}

cloudProvider:
  none: {}
  external: false

addons:
  enable: true
  # In case when the relative path is provided, the path is relative
  # to the KubeOne configuration file.
  path: "../../addons"

apiEndpoint:
  host: '{{ .APIEndpoint }}'
  port: 6443

controlPlane:
  hosts:
{{- $privateKey := "./private.pem" }}
{{- range $value := .Nodes }}
{{- if ge $value.NodeType 1}}
  - publicAddress: '{{ $value.Public }}'
    privateAddress: '{{ $value.Private }}'
    sshUsername: root
    sshPrivateKeyFile: '{{ $privateKey }}'
    hostname: '{{ $value.Name }}'
    taints:
    - key: "node-role.kubernetes.io/master"
      effect: "NoSchedule"
{{- end}}
{{- end}}

staticWorkers:
  hosts:
{{- range $value := .Nodes }}
{{- if eq $value.NodeType 0}}
  - publicAddress: '{{ $value.Public }}'
    privateAddress: '{{ $value.Private }}'
    sshUsername: root
    sshPrivateKeyFile: '{{ $privateKey }}'
    hostname: '{{ $value.Name }}'
{{- end}}
{{- end}}

machineController:
  deploy: false
