# μrl - micro URL Shortener

<p align="center">
  <img src="webpage/images/murl-logo.jpg" alt="μurl-logo" width="120" />
</p>

A minimal URL shortener written in Go that uses Redis to store data. In its current form, it can shorten URLs, resolve URL redirects, and provide information about shortened URLs.

## Technologies Used :fireworks:

- Golang (including gorilla/mux, viper and zap modules):heavy_check_mark:
- Redis :heavy_check_mark:
- Docker :heavy_check_mark:
- Kubernetes :heavy_check_mark:
- Github Actions :heavy_check_mark:

## Table of Contents :round_pushpin:

- [Key Features](#key-features)
- [Installation](#installation)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose)
  - [Kubernetes](#kubernetes)
- [Configuration](#configuration)
- [Testing](#testing)
- [Usage - API calls](#usage)
  - [/ (GET)](#get)
  - [/short (POST)](#short-post)
  - [/info/{shortURL} (GET)](#infoshorturl-get)
  - [/{shortURL} (GET)](#shorturl-get)
- [Future work](#future-work)


## Key Features :point_up:

- Allows users to create custom short URLs
- Generates 6-digit short URL keys if the user does not provide a custom key
- Links, if not used, expire by default in 24 hours (configurable)
- Graceful server shutdown
- Logging, including status and service time for each request
- Implemented as a RESTful JSON API
- Decoupled back-end
- Dockerized


## Installation :coffee:

### Key dependencies-modules used in this project

- gorilla/mux - routing
- uber-go/zap - logging
- spf13/viper - configuration



### Docker
1. Create a Docker network and start Redis container

   ```
   docker network create local-net
   docker run --name redis --network local-net -d redis redis-server
   ```

2. Start the app container

    You can build and run the container image `or` just use the most recently published container image

    - `Option 1` : Build and run the container image

      i.  Clone the repo

      ii. Build and run the app container

    
         ```
         make build-run-image
         ```

      - Note! By default, the app uses the configuration specified in [config-default.yaml](config-default.yaml) and the service listens on port 80. To change the configuration, you can edit [config-default.yaml](config-default.yaml) or use another config file by updating [.env](.env).



    - `Option 2` : Use the latest public container image

      ```
      docker run -it --name url-shortener --network=local-net -p 80:80 mavridis/url-shortener
      ```




### Docker Compose

1. Clone the repo
2. Run the Redis container, and build and run the url shortener container

    ```
    make up
    ```


3. Stop containers and remove previously created containers, networks, volumes, and images

    ```
    make down-del-vol
    ```

    - Note! By default, the app uses the configuration specified in [config-default.yaml](config-default.yaml) and the service listens on port 80. To change the configuration, you can edit [config-default.yaml](config-default.yaml) or use another config file by updating [.env](.env)

### Kubernetes
1. Clone the repo
2. You will need an up and running Kubernetes cluster. You can run the service locally using [minikube](https://minikube.sigs.k8s.io/docs/start/)

    Create Kubernetes resources and run the app

    ```
    make kube-up
    ```

3. A service that creates a load balancer will be used to expose our application. You can list all the services and find out how to access the url shortener service with the following command 

    ```
    kubectl get svc
    ```
    The output should be similar to the following. In this case, you could access the service on 127.0.0.1:80

    ```
      NAME            TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
      redis           ClusterIP      10.101.220.135   <none>        6379/TCP       13m
      url-shortener   LoadBalancer   10.107.46.23     127.0.0.1     80:32099/TCP   13m
    ```
4. Delete the Kubernetes resources created in the first step

    ```
    make kube-down
    ```



## Configuration :pencil:

Checks the [.env](.env) file for the `CONFIG_FILE` environment variable. If `CONFIG_FILE` is empty, then it uses the [config-default.yaml](config-default.yaml) configuration file.

The variables in a configuration file are described as follows:

| Field           | Description                                                                                         | Default Value  |
| ------          | -----------                                                                                         | -------        |
| `server`          |                                                                                                     |                |
| `address`         | string : URL of our service                                                                         | `127.0.0.1:80` |
| `timeoutWrite`    | time.Duration :  The maximum duration before timing out writes of the response                      | `15s`          |
| `timeoutRead`     | time.Duration : The maximum duration for reading the entire request, including the body.            | `15s`          |
| `timeoutIdle`     | time.Duration : The maximum amount of time to wait for the next request when keep-alive is enabled. | `60s`          |
|                   |                                                                                                     |                |
| `redis`           |                                                                                                     |                |
| `address`         | string : URL of database instance                                                                   | `redis:6379`   |
| `pass`            | string : Redis instance password                                                                    | `""`           |
| `database`        | int : Redis database identification                                                                 | `0`            |
| `expiry`          | time.Duration : TTL of each shorted URL                                                             | `24h`          |
|                   |




## Testing :collision:

Service test using the go test command and docker-compose
```
make test
```
And the output should be similar to the following:
```
...
url-shortener-test  | === RUN   TestShortenUrl
url-shortener-test  | --- PASS: TestShortenUrl (0.01s)
url-shortener-test  | === RUN   TestShortenUrlInvalidUrl
url-shortener-test  | --- PASS: TestShortenUrlInvalidUrl (0.00s)
url-shortener-test  | === RUN   TestShortenUrlShortServerURL
url-shortener-test  | --- PASS: TestShortenUrlShortServerURL (0.00s)
url-shortener-test  | === RUN   TestInfo
url-shortener-test  | --- PASS: TestInfo (0.01s)
url-shortener-test  | === RUN   TestInfoURLNotFound
url-shortener-test  | --- PASS: TestInfoURLNotFound (0.00s)
url-shortener-test  | === RUN   TestResolveURL
url-shortener-test  | --- PASS: TestResolveURL (0.01s)
url-shortener-test  | === RUN   TestResolveURLNotFound
url-shortener-test  | --- PASS: TestResolveURLNotFound (0.00s)
url-shortener-test  | === RUN   TestResolveUrlhome
url-shortener-test  | --- PASS: TestResolveUrlhome (0.00s)
url-shortener-test  | PASS
url-shortener-test  | ok 

...
```



## Usage :muscle:
  - ### **/** (GET)

    Returns a static html page that provides basic information about the service


  - ### **/short** (POST)

    Shortens the URL provided and returns a Json response to the user

    - #### **Example WITH user-defined short key**

      Request  :arrow_right:
      ```
        curl -X POST http://127.0.0.1:80/short -H 'Content-Type: application/json' -d '{
          "url":"http://www.testpage.com", 
          "short":"m1"
          }'
      ```
      Response  :arrow_left:
      ```
      {"url":"http://www.testpage.com","short":"m1","expires_in_seconds":86400}
      ```

      Log created on the server :arrow_up_small:
      ```
      url-shortener  | {"level":"info","ts":"2022-06-01T14:52:34Z","caller":"logger/logger.go:28","msg":"Request received","method":"POST","url":"/short","duration":0.0015487,"status":200}
      ```




    - #### **Example WITHOUT user-defined short key**
      Request :arrow_right:
      ```
        curl -X POST http://127.0.0.1:80/short -H 'Content-Type: application/json' -d '{
          "url":"http://www.anothertestpage.com"
          }'  
      ```
      Response  :arrow_left:
      ```
      {"url":"http://www.anothertestpage.com","short":"ec9626","expires_in_seconds":86400}

      ```
      Log created on the server :arrow_up_small:
      ```
      url-shortener  | {"level":"info","ts":"2022-06-01T14:54:19Z","caller":"logger/logger.go:28","msg":"Request received","method":"POST","url":"/short","duration":0.0017039,"status":200}
        ```


  - ### **/info/{shortURL}** (GET)

    Returns information about the specified shortened URL

    Example request  :arrow_right:
    ```
      curl http://127.0.0.1:80/info/m1 
    ```
    Response  :arrow_left:
    ```
    {"url":"http://www.testpage.com","short":"m1","expires_in_seconds":85749}
    ```

    Log created on the server :arrow_up_small:
    ```
    url-shortener  | {"level":"info","ts":"2022-06-01T15:03:26Z","caller":"logger/logger.go:28","msg":"Request received","method":"GET","url":"/info/m1","duration":0.0013197,"status":200}
    ```


  - ### **/{shortURL}** (GET)

    Redirects user to the original URL

    Example request  :arrow_right:
    ```
    curl -v http://127.0.0.1:80/m1 
    ```

    Client is redirected. Detailed response  :arrow_left:
    ```
    ...
       < HTTP/1.1 308 Permanent Redirect
       < Content-Type: application/json; charset=utf-8
       < Location: http://www.testpage.com
       < Date: Wed, 01 Jun 2022 15:16:22 GMT
       < Content-Length: 0
    ...
    ```

    Log created on the server :arrow_up_small:
    ```
    url-shortener  | {"level":"info","ts":"2022-06-01T15:16:22Z","caller":"logger/logger.go:28","msg":"Request received","method":"GET","url":"/m1","duration":0.0014782,"status":308}
    ```  




## Future work :clap:
- Monitoring
- Authentication
