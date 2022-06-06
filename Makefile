default: run

build-run-image:
	docker build -t mavridis/url-shortener -f Dockerfile .
	docker run -it --name url-shortener --network=local-net -p 80:80 mavridis/url-shortener

run:
	docker-compose -f docker-compose.yml up --build 

up:
	docker-compose -f docker-compose.yml up -d --build 

stop:
	docker-compose -f docker-compose.yml stop

down:
	docker-compose -f docker-compose.yml down	

down-del-vol:
	docker-compose -f docker-compose.yml down --volumes

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down --volumes

kube-up:
	kubectl apply -f kubernetes/.

kube-down:
	kubectl delete -f kubernetes/.