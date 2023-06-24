# kubectl-multilog
Stream tailed logs from multiple kubernetes pods
CLI tool for real-time streaming of logs from multiple Kubernetes pods. 

It allows users to monitor logs across containers, across pods and kubernetes clusters, with options for tailing and filtering, useful for debugging purposes.

## Features
* Real-time streaming of logs from selected pods and containers.
* Fetch logs across multiple clusters
* Color-coding to differentiate logs from different pods/containers.
* ~~Fetch logs based on timestamps (sinceSeconds or sinceTime).~~
* Supports logs from init containers.

## Building

To build `kubectl-multilog`, you need Go installed on your machine. You can then build the binary using the following command:

```bash
go build -o kubectl-multilog
```

Move the binary to a directory in your `PATH`, you might need to run this with sudo
```bash
mv kubectl-multilog /usr/local/bin
```


## Usage
Here's how to use the :

``kubectl multilog --help``

```
--kubeconfig <kubeconfig_path> --context <kubernetes,context> --namespace <namespace> --selector <label_selector> --init-containers <bool> --previous <bool>
```

Options:

`kubeconfig:` Path to the kubeconfig file. Defaults to ~/.kube/config.

`context:` Optional, seperated by commas

`namespace:` Kubernetes namespace to use. Defaults to default.

`selector:` Label selector to filter the pods.

`init-containers:` Whether to include init containers. Defaults to false.

`previous:` Whether to include previous terminated containers. Defaults to false.

`container:` Specific container to fetch logs from within a pod.

~~`since:` Fetch logs since a specific point in time or duration (e.g., 5m for last five minutes).~~


## Contributing
Contributions are welcome! Please submit a pull request or create an issue to contribute to this project.

## License
This project is licensed under the [MIT License](https://github.com/umegbewe/kubectl-multilog/blob/main/README.md).

