To build:

```
make build
```

Few examples.

To list all tests that passed in a given run

```
 ./bin/e2e_result show results --passed  --run=2927
I0429 11:14:34.698196    3930 main.go:23]  "msg"="e2e_result tool"  
I0429 11:14:34.787135    3930 utils.go:78]  "msg"="Filter by result:passed"  
I0429 11:14:34.787211    3930 utils.go:97]  "msg"="Filter by run:2927"  
I0429 11:14:34.812837    3930 utils.go:114]  "msg"="Query took 3 milliseconds\n"  
+-------------+------+--------------------------------------------+--------+-----------+
| ENVIRONMENT | RUN  |                    TEST                    | RESULT | DURATION  |
+-------------+------+--------------------------------------------+--------+-----------+
| vcs         | 2927 | sveltos_cluster_provisioned                | passed |  2.251590 |
+-------------+------+--------------------------------------------+--------+-----------+
| vcs         | 2927 | workload_pod_to_node                       | passed |  0.230284 |
+-------------+------+--------------------------------------------+--------+-----------+
| vcs         | 2927 | storage_node_to_NodePort                   | passed |  0.201644 |
+-------------+------+--------------------------------------------+--------+-----------+
| vcs         | 2927 | storage-to-control_node_to_LoadBalancerIP  | passed |  0.236984 |
+-------------+------+--------------------------------------------+--------+-----------+
| vcs         | 2927 | workload-to-storage_pod_to_node            | passed |  0.308979 |
+-------------+------+--------------------------------------------+--------+-----------+
```

To list all runs for which results were collected

```
./bin/e2e_result show runs                        
I0429 11:15:36.182115    4034 main.go:23]  "msg"="e2e_result tool"  
+-------------+------+
| ENVIRONMENT | RUN  |
+-------------+------+
| vcs         | 2927 |
+-------------+------+
| vcs         | 2929 |
+-------------+------+
| vcs         | 2916 |
+-------------+------+
| vcs         | 2917 |
+-------------+------+
| vcs         | 2890 |
+-------------+------+
| vcs         | 2892 |
+-------------+------+
| vcs         | 2898 |
+-------------+------+
| vcs         | 2900 |
```
