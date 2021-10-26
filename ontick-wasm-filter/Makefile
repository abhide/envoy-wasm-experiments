NAMESPACE=ingress
WASM_FILE=ontick-wasm-filter.wasm

build:
	tinygo build -o ${WASM_FILE} -scheduler=none -target=wasi ./main.go

kindly-deploy:
	kubectl create namespace ${NAMESPACE} || true
	kubectl delete configmap wasm-filter -n ${NAMESPACE} || true
	kubectl create configmap wasm-filter --from-file=wasm-filter.wasm=${PWD}/${WASM_FILE} -n ${NAMESPACE}
	kubectl delete configmap envoy-config -n ${NAMESPACE} || true
	kubectl create configmap envoy-config -n ${NAMESPACE} --from-file=envoy.yaml=./envoy.yaml
	kubectl apply -f ${PWD}/deployment.yaml -n ${NAMESPACE}

clean-ns:
	kubectl delete namespace ${NAMESPACE} || true