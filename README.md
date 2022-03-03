# `kubernetes-guestbook`

This is a frontend copy of the Kubernetes Guestbook official application, with a few tweaks compared to the original version:

* **It does not use Angular for the frontend** and instead, uses VueJS 3.
* **It does not come with Redis bundled** and instead, it allows you to use different "storage backends":
  * **Redis** via a third party host, by setting `$REDIS_HOST` with the format `host:port`, and `$REDIS_PASS` in the environment.
  * **MSSQL Server** via a third party host, by setting `$MSSQL_CONNSTRING` in the environment. This must be [a supported connection string from its driver](https://github.com/denisenkom/go-mssqldb#the-connection-string-can-be-specified-in-one-of-three-formats). The quick option is to use a connection string like this: `sqlserver://username:password@hostname:1433`.
* **It uses a key mechanism to have multiple, different copies of the Guestbook** in a single environment.
  * Set a key by setting the `$KEY` environment variable. Be aware that different "storage backends" will have different validations for key names. Currently, both Redis and MSSQL require alphabetic characters only (no numbers, all lowercase).
* **There are multiple storage backends**, and you can choose them by setting the corresponding environment variables for each. On initialization, a `Bootstrap()` function is called, which is often used to connect to the database (and as such, validate connectivity) as well as boostrap any required environment setting.
  * Redis bootstrap ensures there's a key with an empty value
  * MSSQL Server bootstraps creates a database called `cache` and a table with the name of the `$KEY` environment variable
  * Due to how storage backends work, you might need to work around how you want to save the data in the correspondent backend. For example, Redis might use its K/V store, while MSSQL might use a table with a single record which is constantly updated.

PRs are welcome to add new storage backends!

### Use

As a prerequisite, you need a preexistent storage backend. Currently supported: Redis, MSSQL Server.

Creating one in Kubernetes is out of the scope of this document. As an example though, the [`docker-compose.yml` file](docker-compose.yml) will show you its usage with different database backends.

#### Helm chart

In order to use this Helm chart, you can install the repository first:

```
helm repo add guestbook https://patrickdappollonio.github.io/kubernetes-guestbook/
```

You can then install `guestbook/guestbook`, or even render it locally:

```
helm install my-release guestbook/guestbook
```
```
helm template guestbook/guestbook
```

Some of the configuration options you can pass are documented in the [`charts/guestbook/values.yaml` file](charts/guestbook/values.yaml). Specifically, you must configure a backend, either Redis or MS SQL Server.

#### Kubernetes Deployment

The following example shows how to deploy the application to a Kubernetes cluster and use Redis as the storage backend. For other backends, you might need to adjust the `ConfigMap` or `Secret` used.

<details>
<summary>Example generic code</summary>

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: guestbook-config
data:
  redis-host: "redis-master:6379"

---

apiVersion: v1
kind: Secret
metadata:
  name: guestbook-auth
stringData:
  redis-password: "covfefe"

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: guestbook
      tier: frontend
  template:
    metadata:
      labels:
        app: guestbook
        tier: frontend
    spec:
      containers:
      - name: php-redis
        image: ghcr.io/patrickdappollonio/kubernetes-guestbook:latest
        env:
          - name: REDIS_HOST
            valueFrom:
              configMapKeyRef:
                name: guestbook-config
                key: redis-host
          - name: REDIS_PASS
            valueFrom:
              secretKeyRef:
                name: guestbook-auth
                key: redis-password
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        ports:
        - containerPort: 80
```
</details>

In case you want to ensure you have some feedback if the application cannot connect to Redis, you can use an `initContainer`:

<details>
<summary>Example code with init container</summary>

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: guestbook-config
data:
  redis-host: "redis-master:6379"

---

apiVersion: v1
kind: Secret
metadata:
  name: guestbook-auth
stringData:
  redis-password: "covfefe"

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: guestbook
      tier: frontend
  template:
    metadata:
      labels:
        app: guestbook
        tier: frontend
    spec:
      initContainers:
      - name: wait-for
        image: ghcr.io/patrickdappollonio/alpine-utils:latest
        command:
          - sh
          - -c
          - "wait-for-it -w ${REDIS_HOST} -t 300"
        env:
          - name: REDIS_HOST
            valueFrom:
              configMapKeyRef:
                name: guestbook-config
                key: redis-host
      containers:
      - name: php-redis
        image: ghcr.io/patrickdappollonio/kubernetes-guestbook:latest
        env:
          - name: REDIS_HOST
            valueFrom:
              configMapKeyRef:
                name: guestbook-config
                key: redis-host
          - name: REDIS_PASS
            valueFrom:
              secretKeyRef:
                name: guestbook-auth
                key: redis-password
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        ports:
        - containerPort: 80
```
</details>

Be aware the example above cannot validate the Redis password, since the `initContainer` uses a TCP check connection with no understanding of the Redis protocol. For more information, see [`patrickdappollonio/alpine-utils`](https://github.com/patrickdappollonio/alpine-utils/).

#### Docker

The image is hosted in Github Container Registry. [See here for details and versions released](https://github.com/patrickdappollonio/kubernetes-guestbook/pkgs/container/kubernetes-guestbook).

```bash
docker pull ghcr.io/patrickdappollonio/kubernetes-guestbook:latest
```

### Why making another one?

Mostly because I wanted to test what it would be connecting the same frontend to a hosted Redis instance, for example, one from [Redis Labs](https://redis.com/), [Google Memorystore](https://cloud.google.com/memorystore), [Amazon Redis](https://aws.amazon.com/redis/) and a plethora of others.

Additionally, it's a good 3-tier application for easy testing of database connectivity and behaviour.
