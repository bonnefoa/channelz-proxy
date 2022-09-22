# Channelz Proxy

A proxy to call channelz endpoints.

## Install channelz-proxy k8s deployment

To deploy `channelz-proxy`

From the [k8s directory](https://github.com/bonnefoa/channelz-proxy/tree/main/k8s):
```shell
helm template --namespace myns --set toleration=aToleration . | kubectl apply -f -
```

