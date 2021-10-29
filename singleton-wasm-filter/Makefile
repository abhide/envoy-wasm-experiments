NAMESPACE=ingress
PRODUCER_FILE=producer-wasm-filter.wasm
CONSUMER_FILE=consumer-wasm-filter.wasm

build-producer:
	cd producer; tinygo build -o ${PRODUCER_FILE} -scheduler=none -target=wasi ./main.go

build-consumer:
	cd consumer; tinygo build -o ${CONSUMER_FILE} -scheduler=none -target=wasi ./main.go

build: build-producer build-consumer

kindly-deploy:
	kubectl create namespace ${NAMESPACE} || true
	kubectl delete configmap producer-wasm-filter -n ${NAMESPACE} || true
	kubectl create configmap producer-wasm-filter --from-file=${PRODUCER_FILE}=${PWD}/producer/${PRODUCER_FILE} -n ${NAMESPACE}
	kubectl delete configmap consumer-wasm-filter -n ${NAMESPACE} || true
	kubectl create configmap consumer-wasm-filter --from-file=${CONSUMER_FILE}=${PWD}/consumer/${CONSUMER_FILE} -n ${NAMESPACE}
	kubectl delete configmap envoy-config -n ${NAMESPACE} || true
	kubectl create configmap envoy-config -n ${NAMESPACE} --from-file=envoy.yaml=./envoy.yaml
	kubectl apply -f ${PWD}/deployment.yaml -n ${NAMESPACE}

clean-ns:
	kubectl delete namespace ${NAMESPACE} || true
