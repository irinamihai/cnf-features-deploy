package siteConfig

const siteConfigNsErrorCR = `
apiVersion: v1
kind: Namespace
metadata:
  name: ztp-error
  labels:
    name: ztp-error
  annotations:
    argocd.argoproj.io/sync-wave: "0"
`
const siteConfigConfigMapErrorCR = `
kind: ConfigMap
apiVersion: v1
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "1"
  name: ztp-error
data:
`
