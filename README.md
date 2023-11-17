When executed will connect to a k8s cluster and listen for secrets (using the current env's `KUBECONFIG`)

When done monitoring, sending `SIGINT` (Ctrl+C) will trigger analysis of the size of secrets given various 'hardcoded' conditions

```
$ go run .
Collecting Secrets, Ctrl+C to analyze

999 secrets collected ^C

Result:
{
  "NumTotalSecrets": 999,
  "NumTotalPullSecrets": 372,
  "NumSecretsByNsHost": 217,
  "NumSecretsByNsHostWithNoSA": 217,
  "NumSecretsByNsNameHost": 1110,
  "NumSecretsByNsNameHostWithNoSA": 222,
  "SizeBytesByNsHost": 219129,
  "SizeBytesByNsHostWithNoSA": 216297,
  "SizeBytesByNsNameHost": 1170684,
  "SizeBytesByNsNameHostWithNoSA": 219501
}
```