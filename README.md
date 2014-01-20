A server that answers a do-nothing RPC.

Used to test RPC performance.  Currently get about 57K/sec with one
server and one client.  Creating more client processes increases
throughput, even without changing the overall total of outstanding
requests.