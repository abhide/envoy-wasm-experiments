NAMESPACE=ingress
build:
	tinygo build -o ./wasm-filter.wasm -scheduler=none -target=wasi ./main.go

kindly-deploy:
	kubectl create namespace ${NAMESPACE} || true
	kubectl delete configmap wasm-filter -n ${NAMESPACE}
	kubectl create configmap wasm-filter --from-file=wasm-filter.wasm=wasm-filter.wasm -n ${NAMESPACE}
