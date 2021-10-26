NAMESPACE=ingress
WASM_FILE=simple-wasm-filter.wasm

build:
	tinygo build -o ${WASM_FILE} -scheduler=none -target=wasi ./main.go

kindly-deploy:
	kubectl create namespace ${NAMESPACE} || true
	kubectl delete configmap wasm-filter -n ${NAMESPACE} || true
	kubectl create configmap wasm-filter --from-file=${PWD}/${WASM_FILE}=wasm-filter.wasm -n ${NAMESPACE}
